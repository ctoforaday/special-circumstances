# Red audit — round 4, lens: dark-side and risk

Full re-read of `blue/report.md` (833 lines, post-round-3-corrections) in context, plus
`red/findings.md` (all rounds) and `debate.md` in full for prior-round argument state. Live
re-verification, this pass: `git fetch origin` + `git log --oneline -8 origin/main` shows a
**new commit past the report's `d164ab2` pin**: `42dba2d`
("docs(backlog): docket detector tracks IDs not lineages — regression-chain gaps evade the
judge (live run-3 discovery; supersedes-field fix for run 4)", committed
2026-07-14T00:49:14-07:00, 25 minutes after `d164ab2`). `git diff d164ab2..42dba2d --stat` —
`ideas/backlog.md` only, 1 file, docs-only; `debate.js` unchanged. Full direct read of
`debate.js` (219 lines, byte-identical to the `47ae48d`/`88eb57f`/`d164ab2` state prior lenses
verified) confirms the new backlog item's technical claim exactly. Also grepped `debate.md`
(this retrospective's own transcript) for `### LEAD` — **zero matches** across three full
completed rounds.

**Verdict: FAIL.** 1 new gap (R4-1), high-confidence, self-referentially demonstrated by this
exact corpus. Disconfirming pass and checked-but-held items below; none produced a second gap
strong enough to add — this pass concentrates on the one finding because it is structural, not
because the search was shallow (see disconfirming section).

---

## R4-1 — OPEN — HIGH — certain (already realized, 4 times, in this exact corpus) x high x low-medium — corroboration: HIGH (direct read of `debate.js`'s contested-docket logic; live-drift discovery in the project's own backlog naming this retrospective's own gap-id chain as the worked example; direct grep of `debate.md` confirming zero judge invocations across 3 completed rounds)

**Location:** §2.1, Tier A, the round-loop row — *"Round loop / contested docket / deadlock /
safety ceiling / `adjudicated` bookkeeping ... **Correction (round 3, R3-1): this row's 'covered'
framing did not include one degenerate shape**"* — and §2.3 item 3 — *"**Gap-id rollover across
non-adjacent rounds**: `prevGapIds` holds only the prior round, so a gap closed in round 1
recurring in round 3 classifies 'new,' not 'contested'."*

**Problem:** the report's existing gap-id-rollover catch (§2.1 row, §2.3 item 3) is a narrower
special case of a larger, now-confirmed-live defect in the contested-docket detector, and the
report has not yet had the chance to absorb the confirming evidence because it landed after the
round-3 pin.

Direct read of `debate.js` lines 176–190: `const gapIds = new Set(redEnv.gaps.map(g => g.id))`;
`const contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))`; the judge is dispatched only
`if (contested.length > 0)`. This is a pure **id-equality** match. The existing §2.3 item 3
catch assumes the *same* gap id can go missing for one round and reappear — a narrow timing
gap. The defect actually live in this corpus is broader and does not require any round to be
skipped at all: **a gap that is closed-with-regression is, by red's own stated and consistently
applied methodology in this very `red/findings.md`, filed as a *new* gap id** (the fix is
responsive to the original problem but introduces a distinct, smaller defect, graded and IDed
fresh) — e.g. this retrospective's own chain R1-5 → R2-4 → R3-4 + R3-9 (the
`[^DiminishingReturns]` footnote: three consecutive rounds of defects, each given a new id), and
also R2-8 → R3-5, R2-7 → R3-6, R2-5 → R3-10. None of these successor ids ever appear in
`prevGapIds` (which holds only the immediately preceding round's id *set*, by literal string),
so `contested.length` is **zero in every one of this retrospective's three completed rounds** —
confirmed independently by a direct grep of `debate.md` for `### LEAD`: **zero matches**, meaning
the lead-judge has never once been dispatched across this corpus's entire debate history, despite
red's own round-3 disposition explicitly narrating rebuttal-and-regression chains that are the
textbook case the docket exists to route ("legitimate deepening... but legitimate-vs-spinning is
exactly the call the judge exists to make").

This is not a hypothetical extrapolation: `ideas/backlog.md`'s new entry (committed 2026-07-14,
after this report's last pin) independently discovered and named the identical mechanism,
**citing this exact retrospective's own gap-id chain as its worked example** — "red closes gaps
'WITH REGRESSION' and mints successor gaps under fresh IDs (R1-5 → R2-4 → R3-4/R3-9), so
`prevGapIds.has(g.id)` never matches, the contested docket never arms, and the judge never sees a
dispute lineage no matter how long it persists — the only remaining brake is the maxRounds cost
ceiling." Two independent methods (my own control-flow trace + grep, and the project's own live
discovery, filed as a backlog item citing this corpus by its literal gap ids) converge on the
same mechanism — this is stronger corroboration than either alone, in the same pattern the
report's own §1.1 treats as "independent replication, a feature" when lanes converge unprimed.

**Why this is not merely R3-1/R3-2/R3-3 restated:** R3-1 is a degenerate *empty-gaps* FAIL shape;
R3-2 is a dropped friction call site; R3-3 is the judge's carried-gap rationale having no
delivery path *once the judge has been invoked*. All three assume the judge branch is at least
sometimes reachable. R4-1 is one level upstream of all three: it is about the branch never being
entered at all for the corpus's single most common real gap-lifecycle event (repair-regression),
which is a structurally different — and, per the round-2/round-3 shape analysis the report
itself narrates, the *dominant* — defect class in this exact corpus. A fix to R3-1 (empty-gaps
guard) does not touch this; a fix to R3-3 (rationale delivery) is moot if the judge is never
dispatched in the first place.

**Impact on the report's own round-3 forecast:** §4/round-3's closing note — *"if this pattern
holds, round 4 is plausibly the PASS round"* (`debate.md`, Round 3 BLUE) — is not disputed on its
own terms (severity of prose gaps has genuinely declined), but this finding argues the opposite
direction on the *engine* being audited: the mechanism whose entire job is to keep a genuinely
converging debate like this one from silently grinding past its intended arbitration point has
been inert for the debate's whole duration, and the only reason this retrospective happens to be
converging anyway is that blue has, in fact, conceded every gap in good faith each round — a
property of this run's actors, not a property the docket detector enforces. A future run with a
less scrupulous blue (or a genuinely spinning disagreement rather than a genuinely converging one)
would exhibit identical `contested.length === 0` telemetry and grind to `maxRounds` with zero
judge involvement, which is exactly the failure mode `debate.js`'s own comment says the docket
exists to prevent ("so debates converge instead of grinding").

**Required fix:** adopt the backlog's own proposed shape (already scoped, at the merge seat, by
the live discovery — this is not a novel design ask): (1) add a `supersedes: [prior-gap-ids]`
field to the gap object in `RED_ENVELOPE`'s schema; (2) broaden the contested-detection predicate
from `prevGapIds.has(g.id)` to a lineage-chain check (a gap whose `supersedes` list, followed
transitively, reaches depth >= 2 without resolution → route to docket); (3) add a founding-suite
case exercising a 3-round supersession chain (mirroring this retrospective's own
`[^DiminishingReturns]` history) asserting the judge *is* invoked by round 3 of such a chain.
Report-side: correct §2.1's row and §2.3 item 3 to state the broader mechanism (id-equality
alone, not just non-adjacent-round timing, is insufficient) rather than leaving the narrower
framing as the only stated version of this gap. Low-medium complexity: one schema field, one
predicate change, one simulator case — no new infrastructure, matching every other Tier-A fix in
this report's own complexity grading convention.

---

## Disconfirming pass (checked, held — this lens's 1-in-5-equivalent discipline)

- **Does `maxRounds` as a backstop make this an acceptable risk-accept rather than a gap?**
  Checked and rejected: the backstop bounds *cost* (a bad debate cannot run forever), but the
  docket's stated purpose is *convergence quality* — routing a legitimate-vs-spinning question to
  a judgment seat instead of letting N more rounds of unarbitrated back-and-forth accumulate. The
  corpus already shows the cost side of this: 4 rounds and counting of a single footnote's defect
  history (R1-5, R2-4, R3-4, R3-9) that a working docket would have routed to adjudication after
  its second recurrence. A backstop that only bounds worst-case cost, with no working
  intermediate escalation, is not the same control as the one `debate.js`'s own comment claims to
  implement — this is a real gap, not a redundant restatement of an already-accepted tradeoff.
- **Is this just a rephrasing of §2.3 item 3 (already-known gap-id rollover)?** Checked directly
  against item 3's own wording — it is scoped to *the same id* skipping a round, a narrower timing
  case. R4-1 requires no skipped round and no reused id; it is strictly broader and is the one
  actually demonstrated live in this corpus three times over. Kept as a distinct, not
  duplicate, finding — but flagged the overlap explicitly above so blue's fix can retire both
  with one mechanism rather than two.
- **Second candidate gap considered and dropped:** whether `adjudicated` (an array, never
  deduplicated or bounded) could accumulate stale entries across a very long debate and slow the
  `redEnv.gaps.filter` calls. Checked: `adjudicatedIds` is a `Set` rebuilt fresh from `adjudicated`
  each round (line 192) — O(1) lookup, no correctness issue, and at this corpus's gap volume
  (dozens, not thousands) no realistic performance concern either. Not raised — interesting is not
  the same as of interest.
- **Second candidate gap considered and dropped:** the `bulk`/`judgment` model split (row 16b,
  already dispositioned rounds 1–2) re-examined for a security angle — could a cheap bulk-tier
  lens model be more susceptible to being steered by injected content in a poisoned source (§3 row
  19)? No evidence found either way in this corpus; would require a live incident or a
  model-capability citation neither of which exists here. Speculative beyond what the corpus
  supports — not raised as a graded gap, noted here so the question isn't silently unconsidered.

## Verified clean this round

`debate.js` byte-identical across `47ae48d`/`88eb57f`/`d164ab2` and the newly-fetched `42dba2d`
(docs-only backlog commit); the R3-1 (`{verdict:'FAIL', gaps:[]}`), R3-2 (`takeFriction` call-site
count), and R3-3 (judge `carried`-rationale non-delivery) findings all re-confirmed live, unchanged,
at the current HEAD — consistent with round 3's own framing that these are docketed-not-yet-shipped
for run 4, not regressions. `hasNew`/`gapIds`/`contested` computation re-traced line-by-line against
the live script rather than trusted from the round-3 record.

## Friction

None this round for the write path: this payload was produced via the scratchpad-then-copy
convention per this corpus's own documented write-block pattern (filename-keyed, semantic
"report"/"findings" names), as a precaution — `round-4-lens-5.md` does not match that pattern, so
no new block occurrence is claimed or ruled out either way. One process note: confirming R4-1
required `git fetch origin` mid-pass (the live repo advanced by one commit between the round-3
merge and this lens dispatch) — the corpus's own [^PinnedRepoState] discipline is exactly what
caught it; no tooling gap, this is the discipline working as designed.
