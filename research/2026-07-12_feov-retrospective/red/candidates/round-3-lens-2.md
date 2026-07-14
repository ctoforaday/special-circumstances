# Red lens pass — round 3, instance 2 of 3 (slice: §2 Testing strategy + §3 What should change before run 4)

Scope note: sections divided evenly across 3 instances; this pass covers §2 ("Testing strategy:
the trimodal classification and the simulator") and §3 ("What should change before run 4, graded
likelihood x impact x complexity"). Full living report re-read in context (all 770 lines), not
just `blue/CHANGELOG.md`'s Round 2 diff. Per the citation-ledger skip-clause, claims already
verified HIGH in a prior round and not touched by Round 2's CHANGELOG were not re-fetched; claims
in sections Round 2 changed, plus one pure-arithmetic self-consistency check that needs no fetch,
were checked fresh.

## Finding 1 — R2-8's own "reconciliation" arithmetic under-counts by (at least) one lane [HIGH confidence, new this round]

**Location:** §3, row 6, "Reconciled round 2 (R2-8)" clause.

**Quoted text:** "four named methods (primary-literature / practitioner-production /
adversarial-disconfirming-first / local-repo critical-stance) plus a 2-of-N floor on one of them
arithmetically needs `lanes >= 4`... item 6 requires `lanes >= 4` whenever the full four-method
roster with a redundancy floor is in force."

**The problem:** row 6's own redundancy-floor clause (added R1-16, unchanged this round) reads:
"assign the critical-stance/adversarial-disconfirming lens to at least 2 of N lanes (not
1-of-N)" — naming the floored category as a single merged lens ("critical-stance/adversarial-
disconfirming"). But the roster R2-8 itself just enumerated one clause earlier lists FOUR
distinct slash-separated items: primary-literature, practitioner-production,
adversarial-disconfirming-first, AND local-repo critical-stance — i.e., "adversarial-
disconfirming-first" and "local-repo critical-stance" are named as two separate methods, not one.
Doing the arithmetic literally as stated: 3 non-floored methods x 1 lane each + 1 floored method
x 2 lanes = 5 lane-assignments minimum, not 4. The "lanes >= 4" conclusion only holds if
"adversarial-disconfirming-first" and "local-repo critical-stance" are silently treated as the
*same* method for allocation purposes (making the true roster 3 methods, one floored to 2 =
4) — a merge that is asserted implicitly by the floor clause's phrasing but never stated as a
roster-collapsing move, and contradicted by the same paragraph's own "four named methods" count.
This is the identical failure class flagged in my persistent memory
(`pattern_unreconciled_numeric_floors.md`, itself written from this retrospective's round 2): a
floor and an allocation requirement over the same shared resource, individually reasonable,
whose composition was asserted rather than computed — except here the "reconciliation" that was
supposed to close exactly that gap (R2-8) reintroduces a version of it under a corrected-looking
number.

**Compounding issue, same row:** the reconciliation also asserts "row 7 below floors N at 3 (the
shipped default)" — but row 7 itself, two rows down and unchanged this round, is graded
**[OPEN]**: "confirmed still absent: `lanes = 3` remains an unguarded default with no minimum
check on `main`." A default value that nothing prevents overriding downward is not a floor — and
the report's own §1.1 opening line documents a live case of exactly that override (run 2 shipped
with `--lanes 2`). Calling row 7 "the shipped default" floor in the same sentence that computes a
lane-count requirement is describing an enforcement mechanism that does not exist yet, which
matters because the whole point of row 6's reconciliation is to state a hard minimum operators
must respect.

**Verification method:** pure arithmetic / internal-consistency check against the report's own
stated roster (no external fetch required — this is a leaf-node check where the "source" is the
report's own prior paragraph).

**Risk grade:** likelihood — certain, this is a static logic property, not a live-system
observation, so it will misfire identically on every future read; impact — medium: a future
operator implementing "item 6's full roster with a redundancy floor" at the literally-stated
`lanes >= 4` floor will silently drop one of the four named methods (most likely reintroducing
the exact failure-concentration risk R1-16 was raised to close, this time via a wrong-by-one
floor rather than an absent one); complexity to fix — trivial: either (a) state the true minimum
as `lanes >= 5` if all four methods are meant to be distinct, or (b) state explicitly that
"adversarial-disconfirming-first" and "local-repo critical-stance" collapse into one lens for
allocation purposes (reducing the named roster to 3 for this arithmetic), and align row 7's
"shipped default" language with its own [OPEN] status.

## Finding 2 — R2-7's grep citation is imprecise: the "one exception" it implies does not exist [LOW-severity, precision only]

**Location:** §3, row 19 (content-poisoning risk-accept, "rewritten round 2, R2-7").

**Quoted text:** "a repo-wide grep for 'independent' in the plugin returns zero hits outside this
ledger clause's own text."

**Verification:** `git grep -ni "independent" -- plugins/frank-exchange-of-views` (live, this
round, `main` @ `d164ab2`) returns **zero matches, full stop** — including inside `debate.js`'s
`ledgerClause` string (line 156, read directly: "CITATION LEDGER: read .../citation-ledger.md
first if it exists; a claim verified at HIGH confidence in a prior round stays verified — do not
re-fetch it unless .../blue/CHANGELOG.md shows its section changed this round. Append every claim
you verify to the ledger..." — no occurrence of "independent" anywhere in that string either).
So the sentence's literal claim ("zero hits *outside* this ledger clause['s text]", implying at
least one hit *inside* it) is not what a repo-wide grep actually shows; the accurate statement is
simply "zero hits in the plugin, period" — which is if anything *stronger* evidence for row 19's
underlying point (the protocol truly has no "independent"-cross-referencing language anywhere,
not even in the one place a reader might expect it), but the sentence as written sends a future
verifier looking for an exception that isn't there.

**Risk grade:** likelihood a reader is misled — low (the substantive disposition is unaffected
and the correction only strengthens it); impact — low (no verdict or disposition depends on the
precise wording); complexity to fix — trivial (drop "outside this ledger clause's own text," or
replace with "returns zero hits, full stop").

## Re-verified this round, no new gap (Round-2-changed sections in this slice, confirmed still accurate)

- **§3 row 6 / §1.1 cross-provider citation (R2-3):** arXiv:2602.03794 Table 2 distinction
  (L4 2-agent 67.71% vs. L1 16-agent 65.34%; L2 8-agent 65.44%) — not re-fetched (already HIGH,
  merge-time re-fetch, round 2 ledger); consistent with the current report text.
- **[^DiminishingReturns] footnote (R2-4 rebuttal of red's own round-2 proposed citation):**
  independently re-fetched arXiv:2606.02646 abstract + HTML this round (not merely trusting
  blue's self-check of a rebuttal against red's own prior proposal, since that is exactly the
  kind of self-grading this lens exists to catch). Confirmed accurate at HIGH confidence: the
  paper's benchmarks are MMLU-Hard / **GSM-Hard** (not GSM-Plus) / GPQA Diamond; it states "the
  practical knee is N≈10"; and "N_eff saturates near 1.8 on GSM-Hard" (free-form math) — matching
  blue's rebuttal almost verbatim. Blue's rejection of red's proposed source and the dropped
  "7 agents" clause both hold up under independent re-fetch.
- **§3 row 4 (R2-11, `--smoke`):** confirmed live — exact string `--smoke` has zero matches in
  `commands/research.md` and exists only in `debate.js`'s header comment (lines 17-18,
  descriptive, not parsed). Note: `commands/research.md` line 9 does say "haiku for smoke tests"
  in prose, meaning an operator can already manually approximate a smoke run today via existing
  `lanes`/`maxRounds`/`model` parameters — consistent with row 4's own complexity grade ("threads
  existing tunables"), so this is not a contradiction, just confirms the framing is accurate as
  written.
- **§3 row 10 (R2-9, ledger skip-trigger):** `debate.js` line 156 (`ledgerClause`) unchanged
  between `47ae48d`/`88eb57f` and current `main` @ `d164ab2` (`git diff` empty on this file across
  all three refs) — the prose-change-keyed skip-trigger description still matches the shipped
  code exactly.
- **§3 row 11 / §5 items 4, 7 (R2-6, run-3 artifact-trail gap):** `git ls-tree main -d research/`
  still shows exactly two run directories (`2026-07-12_feov-retrospective`,
  `2026-07-12_memory-architecture`) — no run-3 directory, confirmed at current HEAD. Live
  backlog drift since the report's pin (`88eb57f` -> `d164ab2`, one line changed in
  `ideas/backlog.md`, item 28 gained a new sub-item (d) "MERGE-SEAT ANALYSIS (run-3
  transcripts)... red-merge-r1: ~100-150K of material, 2.7M+ cache reads") — this is further,
  independent confirmation that run 3 executed and left transcript-level evidence while
  `research/` still has no run-3 directory. The quoted (b) finding the report's own
  [^CostFigureProvenance] footnote cites is unchanged verbatim; this drift does not contradict
  anything already claimed, it only deepens the existing R2-6/row-11 gap. Not filed as a separate
  new finding since it corroborates rather than contradicts an already-conceded gap — noted here
  so the next audit round doesn't need to rediscover it.
- **§0/§3 row 2, 2b (judge-null, citationPasses-const):** re-confirmed live at `d164ab2` — no
  `if (!judge)` guard exists anywhere in `debate.js`; `citationPasses` (line 139/now still a
  `const`, computed once before the `while` loop) is unchanged. `debate.js` is byte-identical to
  `47ae48d` per direct diff (only `ideas/backlog.md` changed between `88eb57f` and current HEAD).
  No drift affecting these rows.

## Convergence note

Nothing in this slice disputes the H1–H5 verdicts. Both findings this round are precision/
consistency gaps in blue's own round-2 repair work, not new substantive gaps in the underlying
proposals — consistent with the "repair-regression" pattern this retrospective's earlier rounds
already named as the dominant round-2 failure mode. Finding 1 is the more load-bearing of the two
and should block a clean PASS on row 6 until the arithmetic is stated correctly; Finding 2 is a
wording fix only.
