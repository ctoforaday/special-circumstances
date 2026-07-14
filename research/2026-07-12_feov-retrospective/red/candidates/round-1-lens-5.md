# Red audit — round 1, lens: dark-side and risk

Scope: failure modes, likelihood x impact x complexity grading, security and tradeoff blindspots.
Full re-read of `blue/report.md` (554 lines) completed. Leaf-node verification against the live
`special-circumstances` repo (current `main` @ `00018a5`, not the report's verification point
`9ff0fad`), `gh pr view 14`, and two external MCP-server trust checks.

## Verdict: FAIL — 8 gaps, none closing the report but two (G1, G2) invalidate its central
disposition table as currently worded and must be corrected before this report is usable as a
run-4 planning document.

---

### G1 — Live-source drift: the headline's central claim is now false on the current primary source [HIGH likelihood x HIGH impact x LOW complexity]

**Location:** §0 Headline — `"[L3] The claimed fixes are absent from main."` and `"[L1] The
fixes exist — on an open, unmerged pull request."`

PR #14 **merged to `main`** at commit `00018a5`, `2026-07-14T05:58:54Z` — confirmed via
`gh pr view 14 --json state,mergedAt,mergeCommit` (`state: MERGED`) and `git merge-base
--is-ancestor 00018a5 origin/main` (true; it is on the pushed remote, not a local artifact). The
merge landed ~8 minutes after the report's own verification commit (`9ff0fad`, `22:50:27 -0700`
vs. merge `22:58:54 -0700`) — a genuine race, not carelessness — but the report as written now
misleads any reader who opens it today: `debate.js` (renamed, confirmed present), the args
guard, the `blueEnv`/`redEnv` null-guards, and `tests/simulator/` are all on `main` right now, not
"on an open, unmerged pull request." This is a textbook live-source-drift finding (project memory:
`gap_live_source_drift.md`) and it is the load-bearing premise for:
- §0's entire "shipping question, not research question" framing,
- §3 item 1's disposition ("Do first. This is a shipping decision, not a proposal"),
- §3 items 2–3's "subsumed by #1 *only if* #1 merges" conditional (now triggered — see G2),
- §4 rows 3 and 4's status column ("Fixed on PR #14, **unmerged**").

**Required fix:** re-verify against current `main` HEAD and correct every one of the four locations
above; do not just flip "unmerged" to "merged" — see G2, because a naive flip overstates
resolution.

**Corroboration confidence:** high (git + `gh` CLI, direct, reproducible right now).

### G2 — The merge does not close item 2; "subsumed by #1" is not yet true, and a naive headline fix would wrongly imply it is [HIGH x HIGH x LOW]

**Location:** §3 item 2 disposition — `"subsumed by #1 *only if* #1 merges and the suite extends to all call sites"`; §2.3 item 1 — `"Null at every schema'd call site, not two ... A suite that only guards the observed site is not founding — it is anecdotal."`

Direct read of `main`'s `debate.js` (218 lines) confirms: `topic`/`runDir` guard present; `blueEnv`
null-guard present (`if (!blueEnv) throw`); `redEnv` null-guard present. But the **judge call site
has no null-guard**:

```js
const judge = await agent(`Adjudication, round ${round}...`, { ...judgment, ... schema: JUDGE_ENVELOPE })
for (const r of judge.resolutions) { ... }   // <- throws if judge is null
```

A quota-wall or terminal failure on the adjudication call reproduces run 2's exact `TypeError`
crash class, on `main`, today, at a fourth call site the merge did not cover. This is precisely
the "not two" scenario §2.3 item 1 already anticipated in the abstract — but the report was
finalized before the merge and so never checked which sites the actual merged diff covers. If
blue's round-2 revision simply updates G1's four locations to "merged" without re-running this
check, the report will read as fully resolved when it is not.

**Required fix:** add the judge-null-guard (and re-audit final-assembly's unguarded `agent()`
call, though its return is currently unused so it is not a live crash risk); re-word §3 item 2's
disposition from conditional ("only if") to factual ("confirmed still open on `main`: judge call
site").

**Corroboration confidence:** high (direct source read, line-quoted above).

### G3 — `citationPasses` is computed once before the debate loop and never rescales as blue's report grows — a live shipped defect, not just a missing test case [HIGH x MEDIUM-HIGH x LOW]

**Location:** §2.3 item 4 — `"citationPasses recomputed from the new claim_count"` (listed as an
addition to test); §2.1 Tier A — `"citationPasses arithmetic ... Pure arithmetic; boundary table"`.

`debate.js`: `const citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count ||
20) / 40)))` sits immediately after the *initial* blue synthesis, **before** `while (round <
maxRounds)`, and is never reassigned inside the loop — including after the line that reassigns
`blueEnv` from each round's blue-response. Blue's report is additive by design (never subtract
substance; §0's own doctrine); claim count should grow round over round as gaps get addressed
with more claims. Citation-verification capacity does not grow with it: a report that starts at,
say, 35 claims (1 pass) and grows to 180 claims by round 4 still gets exactly 1 citation pass per
round for the rest of the debate — a systematically under-scaled audit exactly in the later
rounds where the report is largest. §2.3 item 4 describes this as a test case to *add*, phrased as
though the recompute already happens ("citationPasses recomputed...") without the "known-failing
until fixed" flag that items 2 and 3 in the same list correctly carry. It does not happen; this is
a code fix, not only a test case, and it has no row of its own in §3's 18-item graded table.

**Required fix:** move the `citationPasses` computation inside the loop, recomputed from the
latest `blueEnv.claim_count` each round; add a §3 row; label the §2.3 item 4 test as
known-failing until then, matching items 2–3's convention.

**Corroboration confidence:** high (direct source read).

### G4 — Per-role model routing cheapens the adversary's actual audit work, contradicting the doctrine it's shipped under [MEDIUM-HIGH x MEDIUM-HIGH x LOW]

**Location:** §3 item 16 — `"Per-role model split ... Ships with #1 ... never change model/judgmentModel on a resume"`.

`debate.js`'s own doctrine comment: *"cheapen redundancy and mechanics, **never judgment or the
adversary**: `model` drives the BULK seats (frontier, blue lanes, **red lenses**, blue responses);
`judgmentModel` drives the JUDGMENT seats (blue-synthesize, **red-merge**, lead-judge,
assemble)."* `ideas/backlog.md` restates the same doctrine ("cheapen redundancy, never
judgment"). But the individual red-lens passes — the actual dispatches that do leaf-node citation
verification, logic auditing, and this dark-side/risk audit — are assigned to the cheap `bulk`
tier; only `red-merge` (the mechanical consolidation of lens outputs into `findings.md`) gets the
`judgment` tier. The stated doctrine's own word "**the adversary**" most naturally names the
audit itself (what a red lens produces), not its consolidation step — yet the routing table
cheapens exactly that. §3 item 16 reproduces the resume-cache caveat but never examines this
tension, so a dev/smoke run with `model=haiku` silently downgrades leaf-node verification quality
(the report's own repeated finding that leaf-node checks are where the load-bearing catches
happen — e.g. §1.4's false-premise repo verification) at precisely the runs (dev, smoke) where the
harness is being validated.

**Required fix:** either (a) reclassify red-lens dispatches to the judgment tier, or (b)
explicitly document that lens-tier cheapening is a deliberate, bounded cost/quality tradeoff
scoped to non-keeper runs, with a stated confidence discount applied to any gap a
cheap-model lens pass raises.

**Corroboration confidence:** high (direct source read of both the code comment and the backlog,
both asserting the same doctrine the routing table contradicts).

### G5 — Item 15's risk-acceptance argument for ENAMETOOLONG doesn't logically hold, and undercounts recurrence [MEDIUM x LOW-MEDIUM x LOW]

**Location:** §3 item 15 disposition — `"Risk-accept pending PR #14's live trial — the skeleton fix (append-not-write) may moot most of it by construction; re-measure after run 3."`

ENAMETOOLONG is a shell-command-length ceiling on a single large heredoc argument; it fires on
the *size of the payload in one call*, independent of whether the target file is fresh (Write) or
pre-existing (Edit/append). The skeleton fix changes which tool-call shape is used
(Write-of-new-file vs. append-to-existing) but does not shrink the payload of a large append —
the same oversized single-shot write would still overflow the command-length ceiling whether it
targets a fresh or pre-created file. The causal claim ("may moot most of it by construction") is
not supported by the mechanism described elsewhere in the same report (§3 item 8, §0 addendum).
Separately, this defect has now recurred **three times** across two runs plus this synthesis
itself (§0's live addendum: "a first chunked-heredoc workaround attempt then failed on shell
parsing"), which argues for at least Medium-High likelihood, not the implicit lower-touch framing
of "risk-accept pending trial."

**Required fix:** either drop the "(a) may moot most of it" clause or replace it with an accurate
mechanism (e.g., a chunked-append helper, independent of the write-block fix); re-grade likelihood
given three recorded occurrences.

**Corroboration confidence:** medium (the logical gap is direct source-comparison within the
report itself; the "three occurrences" count is the report's own claim, not independently
re-verified by this lens beyond what's already footnoted).

### G6 — Adopting third-party MCP servers is graded "Low complexity" with no security-vetting step for external tool code gaining file/network access inside the pipeline [MEDIUM x MEDIUM x LOW]

**Location:** §3 item 13 — `"Low once scoped as adoption: off-the-shelf MCP servers exist (arxiv-latex-mcp ... pdf-reader-mcp ...) — no bespoke sc-pdf-extract Go tool needed."`

Live-checked both: `arxiv-latex-mcp` (takashiishida) shows a third-party security-scan report of
100/100 with no issues found; `pdf-reader-mcp` (SylphxAI) is actively maintained (v2.3.0,
577+ stars, CI, 397 tests) — so this is not a "the tool is untrustworthy" finding; both check out
reasonably well on a live look. The gap is structural: the complexity grade for wiring an
external MCP server into a research pipeline that will feed its output straight into the
corroboration-confidence chain has zero line item for reviewing the server's source, pinning a
version, or scoping its filesystem/network permissions before adoption — the same category of
supply-chain trust question the report is happy to interrogate for CVE-2026-21852 in a sibling
run's subject matter (§1.1) but never applies to its own tooling choices.

**Required fix:** add one sentence to item 13's disposition: pin version, review source/maintainer
before wiring in, scope the MCP server's declared permissions.

**Corroboration confidence:** high for the tools' current trust signals (live web search, dated
2026-07-13/14); the "no vetting step" observation is a direct read of the report's own text (no
independent verification needed — it's an absence).

### G7 — The report never turns its own convergent finding (content-poisoning) on the harness's own fetch surface [MEDIUM x HIGH x MEDIUM]

**Location:** §1.1 — `"CVE-2026-21852 memory poisoning as the 'absent from §9 entirely' blocking omission ... arrived at independently."`

The report devotes real space to a sibling run's finding that an agentic system's context can be
poisoned via untrusted ingested content, and to that sibling run's own subject-matter blind spot
in failing to flag it. FEOV's own `/research` pipeline autonomously fetches untrusted external
content (WebSearch/WebFetch across ~22 searches this run alone) into a multi-agent debate with
live filesystem write authority over the repo — the identical shape of risk (untrusted external
content -> agent context -> downstream action) the report treats as a headline finding elsewhere.
Nowhere in this 554-line report — §0 through §5, all 18 graded rows of §3 — is this risk class
applied reflexively to the FEOV harness's own research phase. This is not a claim that the harness
is currently exploited; no evidence of that exists. It is an absence: a "dark-side and risk" pass
over a document that spends a full subsection on this exact vulnerability class elsewhere should
at minimum note whether it applies here and, if so, grade it.

**Required fix:** add a graded row (or an explicit risk-accept with rationale) addressing whether
fetched web content can influence blue/red output in ways that survive into `report.md` /
`findings.md` without a human in the loop noticing — even a one-line "out of scope, here's why"
closes this more honestly than silence.

**Corroboration confidence:** medium — the parallel is structural/reasoned, not sourced to a
specific exploit; graded as "of interest," not proven live risk.

### G8 — Lane specialization (item 6) trades convergence-reduction for concentration of failure risk, undiscussed [LOW-MEDIUM x MEDIUM x LOW]

**Location:** §3 item 6 — `"one added sentence per lane-dispatch prompt; no synchronization between lanes."`

Today, every lane is assigned "take hypothesis i first, then breadth" — fully redundant in method;
if any one lane's `agent()` call returns null or fails, the others still independently cover
breadth and, per §1.1's own refinement 2, still occasionally converge on the same real gap by
different paths (the CVE-2026-21852 double-catch). Item 6's proposed fix — one lane always
primary-literature, one always practitioner-production, one always
adversarial/local-repo-critical-stance — measurably reduces convergence (its whole point, and the
cited ~19% correlation-reduction figure supports it) but also means a single failed dispatch for
the "critical-stance" lane now drops 100% of that method's coverage for the round, with no other
lane positioned to catch what only that lens catches (per this run's own finding that the
false-premise repo verification in run 2 came from exactly one lane doing exactly one method). The
complexity column notes "no synchronization between lanes" but not this failure-mode trade;
neither §3 nor §5's open questions carries it.

**Required fix:** either a minimal redundancy floor (e.g., two lanes share the
adversarial/critical-stance assignment) or an explicit risk-accept noting the trade and a
re-dispatch-on-null policy for specialized lanes.

**Corroboration confidence:** medium — reasoned from the report's own §1.1 material and the
`debate.js` `parallel()` semantics (a thrown/null lane resolves independently per PR #14's
harness), not independently re-tested.

---

## Disconfirming pass against this lens's own findings

- G1/G2: could the merge have been reverted or the `judge`-guard already patched post-merge in a
  commit this check missed? Re-checked `git log --oneline -3 main` and re-read `debate.js` at HEAD
  directly (not cached) — no, confirmed current.
- G4: could "the adversary" in the doctrine comment intentionally mean only `red-merge`/judge
  seats (the seats that decide verdicts), making the routing self-consistent? Plausible reading,
  noted in the required-fix as option (b) — this is why G4 is graded a documentation/tradeoff gap
  with two possible resolutions, not asserted as an unambiguous bug.
- G6: could the MCP servers already have been vetted by the operator informally, off-corpus? No
  evidence either way; graded as an absence in the *document*, not a claim about undocumented
  operator diligence (which per protocol counts as untrusted/self-report anyway).

## Not raised as gaps (checked, held to a high bar)

- §3 item 14 (advisory-access risk-accept) — argued with evidence (§13.7 CHANGELOG citation), a
  real workaround shipped; accepted.
- §3 item 18 (audit-narrowing hold) — correctly identifies its own failure mode (a missed
  regression) and defers to human gating; accepted as an honest risk-accept, not a soft-pass.
- Write-block layered fix (item 8(a)/(b)/(c)) — already carries its own uncertainty (issue
  #13890, open question 4) at an appropriate grade; not re-litigated here beyond G5's narrower
  point about item 15's distinct mechanism claim.
