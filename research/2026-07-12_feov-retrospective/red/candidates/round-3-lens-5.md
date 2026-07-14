# Red audit — round 3, lens: dark-side and risk

Scope: failure modes, likelihood x impact x complexity grading, security and tradeoff blindspots.
Full re-read of `blue/report.md` (770 lines, post-Round-2-corrections) completed in context, not
from `blue/CHANGELOG.md`'s Round 2 diff summary alone. Also read `red/findings.md` (round 1 and
round 2 sections) and `debate.md` in full for prior-round context and to avoid re-litigating
closed ground.

Leaf-node verification against the live `special-circumstances` repo, this round: `git log
--oneline -5 origin/main` shows `main` has advanced *again*, to `d164ab2` ("docs(backlog):
merge-seat cost analysis..."), one commit past the report's last-verified `88eb57f`. Checked the
diff: `git diff 88eb57f..d164ab2 -- plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js`
is empty (the new commit touches only `ideas/backlog.md`, +1/-1) — `debate.js` is unchanged since
the report's last pin, so no report claim is invalidated by drift this round. Then did a full,
line-by-line direct read of the live `debate.js` (219 lines) specifically hunting for
control-flow/telemetry edge cases the report's existing risk grid does not cover — the leaf-node
citation lenses already exhaustively re-verify the *cited* facts (guards, routing, ledger clause
text); this pass instead traced the script's actual runtime paths for degenerate-but-schema-legal
envelope shapes, since that is exactly this corpus's dominant defect class (H3/H4) and exactly this
lens's mandate.

## Verdict: FAIL — 3 new gaps (R3-1, R3-2, R3-3). None overturn H1–H5, none dispute a round-1/round-2
closure. All three are previously-unexamined runtime/telemetry edge cases in the *already-merged*
`debate.js`, found by tracing control flow rather than re-checking citations. R3-2 directly
falsifies an existing factual claim in the report's own §2.1 table ("never dropped").

---

### R3-1 — A schema-legal but semantically-degenerate red-merge envelope (`verdict: 'FAIL'`, `gaps: []`) silently burns every remaining round to the safety ceiling, then reports a misleading terminal state [LOW-MEDIUM likelihood x MEDIUM-HIGH impact x LOW complexity]

**Location:** §2.3, item 8 — *"`--maxRounds 0` — the emitted log line must distinguish 'never ran'
from 'ran and failed at round 0'"* and §2.1's Tier A row *"Round loop / contested docket / deadlock
/ safety ceiling / `adjudicated` bookkeeping ... Pure `Set`/array logic over canned envelope shapes"*
— both describe the round-loop's degenerate-input handling as covered ground, but neither covers
this specific shape.

**Problem:** `RED_ENVELOPE` (`debate.js` lines 56–91) requires `verdict` and `gaps`, but nothing
constrains `gaps` to be non-empty when `verdict === 'FAIL'` — no `minItems`, no cross-field check.
Direct trace of the round loop (lines 148–198) for this exact input: `gapIds = new
Set(redEnv.gaps.map(...))` → empty Set; `contested = redEnv.gaps.filter(...)` → `[]`; `hasNew` →
`false`. `if (contested.length > 0)` (line 180) is false, so **the judge is never invoked** — there
is no adjudication path for "FAIL with nothing to adjudicate." `prevGapIds` resets to empty.
`openGaps = redEnv.gaps.filter(...)` → still `[]`. Blue is then dispatched (line 194) with the
literal prompt *"Red's verdict: FAIL. Open gaps (adjudicated ones excluded): []."* — blue has
nothing concrete to act on but is told the debate failed. Nothing in the loop detects this
inconsistency; it recurs identically every round until `round >= maxRounds`, at which point
`exhausted` becomes true (line 200) and the run terminates via the safety ceiling — silently, no
distinguishing log line, no thrown error. The final returned envelope (lines 210–218) then reports
`verdict: 'UNVERIFIED'` alongside `gaps_outstanding: redEnv.gaps.length` = **0** — an operator
reading only the top-level return sees "UNVERIFIED, zero gaps," a directly self-contradictory
terminal signal, with no indication that N paid rounds were burned on empty red-merge output. This
is not covered by §2.3 item 10 ("malformed-but-non-null envelope, e.g. `gaps` missing") — that case
is a *schema violation* (missing required field); this case is *schema-valid* (an empty array
satisfies `type: 'array'`) but semantically incoherent, a distinct failure mode from the one already
catalogued.

**Required fix:** one guard after the `redEnv` null-check (line 171-172): if
`redEnv.verdict === 'FAIL' && (!redEnv.gaps || redEnv.gaps.length === 0)`, either treat it as an
effective PASS-with-a-warning (a FAIL with no gaps is, functionally, nothing left to fix) or throw
a distinguishing error rather than looping silently to the ceiling. Add as a 12th/13th simulator
case alongside the existing `--maxRounds 0` case (§2.3 item 8), since both are "degenerate loop
termination" cases of the same family.

**Corroboration confidence:** high — traced directly against the live, unchanged `debate.js` at
`d164ab2` (lines 56–91 schema, 148–198 loop body, 200–218 return); no external citation involved.

---

### R3-2 — Blue-synthesize's (round-0) friction is never harvested into the aggregated `friction` array — directly falsifies the report's own "never dropped" claim, and would swallow the exact live event (§0) this report is proudest of [MEDIUM-HIGH likelihood-already-realized x MEDIUM impact x LOW complexity]

**Location:** §2.1, Tier A table — *"Friction aggregation: per-seat arrays namespaced by label and
concatenated, never dropped | source | Self-improvement input integrity [L1+L3]"* — this claim is
false for one schema'd seat, and the seat in question is the one this very report's own §0 live
addendum documents firing.

**Problem:** `takeFriction(who, env)` (line 146) is called at exactly three sites: `red-merge-r${round}`
(line 170), `judge-r${round}` (line 187), and `blue-respond-r${round}` (line 197). It is **not**
called for the round-0 `blue-synthesize` dispatch (lines 132–136), even though `blue-synthesize` is
schema'd against `BLUE_ENVELOPE` (line 134: `schema: BLUE_ENVELOPE`), and `BLUE_ENVELOPE` itself
declares a `friction` field (line 52: `friction: { type: 'array', items: { type: 'string' } }`) —
the schema supports exactly this signal, and the code silently discards it. This is not a design
choice consistent with "unschema'd seats can't report friction" (which is true and fine for
`frontier`/blue-lane dispatches, which return free-text synopses with no schema at all) — it is an
oversight specific to one schema'd, friction-capable seat. Concretely: this retrospective's own §0
live addendum ("the write-block fired *again* on this very report — the synthesizer's Write of
`blue/report.md` was refused... producing it via scratchpad Write + copy") describes exactly a
round-0 blue-synthesize friction event. If the *already-merged* `debate.js` had been running this
retrospective live, that friction — reported by the agent in its `BLUE_ENVELOPE.friction` field —
would never reach the aggregated `friction` array, never reach the final-assembly prompt's
"Collated friction so far" (line 207), and never reach `/self-improve`'s input. The only reason
this report's §0 addendum exists at all is that the synthesizing agent narrated the event into
prose rather than relying on the structured channel — which is exactly the "self-observed, not
yet artifact-logged" weakness R1-18/R2-2 already downgraded that same evidence for, for an
unrelated reason (no independent corroborator). Here the corpus's own regression suite doesn't
catch this either: the merged test (`tests/simulator/debate.test.mjs`, "friction aggregates from
every seat with attribution") stubs and asserts only `red-merge-r1` and `blue-respond-r1`'s
friction — it does not stub or assert blue-synthesize's, so 11/11 green gives false confidence that
"every seat" is covered when one schema'd, friction-capable seat demonstrably is not.

**Required fix:** add `takeFriction('blue-synthesize', blueEnv)` immediately after the null-guard
at line 136 (guard first, since a null `blueEnv` has no `.friction` to read); add a 13th/14th
simulator case stubbing blue-synthesize's friction field and asserting it appears in the final
`friction` array, mirroring the existing red-merge/blue-respond case but for the one seat it
currently omits. Correct §2.1's "never dropped" claim to name the one exception until the fix
ships (it is currently untrue as a general statement, not "generally true with one edge case
noted" — it reads as an absolute).

**Corroboration confidence:** high — traced directly against the live, unchanged `debate.js`
(lines 132–146, 170, 187, 197) and the merged test file (`git show
feat/feov-dogfood-round-1:.../debate.test.mjs`, lines 114–123, direct read this round) which
confirms the passing suite's actual scope stops at red-merge/blue-respond.

---

### R3-3 — The judge is explicitly instructed to state "what further research blue owes" for a carried gap, but the script has no channel to deliver that rationale to blue's next dispatch [MEDIUM likelihood (fires whenever any gap is carried) x MEDIUM impact x LOW complexity]

**Location:** `debate.js`'s own judge prompt (line 182) — *"Rule per contested gap (closed |
rebuttal_sustained | risk_accepted | carried | unresolved) with rationale — **for carried, state
what further research blue owes**"* — against §2.3 item 5's suite-case description — *"Judge
`carried` branch — deadlock stays false; carried gap re-enters `openGaps` with its required-fix
intact [L3]"*.

**Problem:** the judge is explicitly directed to produce a specific, targeted piece of guidance
("what further research blue owes") when carrying a gap forward — the whole point of the "carried"
resolution, as distinct from just leaving the gap open, is that the judge has something more
precise to say than red's original `required_fix`. But trace the data flow: the judge's structured
return (`JUDGE_ENVELOPE`, lines 93–112) only carries `resolutions[].rationale` back into the
script's `judge` variable; that variable is consumed only to filter into `adjudicated` (lines
184–186, and only for closed/rebuttal_sustained/risk_accepted — "carried" is explicitly excluded
from this filter) and to pull `judge.friction` (line 187). **The rationale text for a `carried`
resolution is read nowhere in the script and passed into no subsequent prompt.** The next round's
`blue-respond` dispatch (line 195) is built entirely from `redEnv.gaps` (red's *original* gap
object — `id, location, problem, required_fix, severity, likelihood, impact, complexity_cost` —
none of which is the judge's rationale) via `openGaps`; the judge's own `debate.md` append (which
the judge prompt does separately instruct) is the only place the "what further research" guidance
survives, and blue-respond's prompt (line 195) never instructs blue to read `debate.md` before
responding. §2.3 item 5's own suite-case description confirms the test only checks that the
*gap's* `required_fix` survives the carry, not that the judge's distinguishing rationale reaches
blue — so the passing test doesn't exercise this path either. Net effect: the system asks the
judge for something specific and valuable, and then has no mechanism to use it — the judge's answer
is written to a file for a human to read, not threaded back into the automated loop it was
requested for.

**Required fix:** either (a) fold the judge's per-gap `rationale` for `carried` resolutions into
the `openGaps` payload passed to `blue-respond` (e.g., attach `judge_rationale` alongside the
matching gap id before building `openGaps`), or (b) add one sentence to the `blue-respond` prompt
instructing blue to read the latest `### LEAD` section of `debate.md` for any carried-gap guidance
before responding. (a) is more robust (survives if blue skips the read); (b) is cheaper (no new
data plumbing) but relies on blue's own initiative, the same class of reliability problem this
corpus has already flagged elsewhere (backlog item 4, blue not reading red's memory library,
§3 row 12).

**Corroboration confidence:** high on the mechanism (direct trace of `debate.js` lines 93–112,
166–197, and the merged test at `debate.test.mjs` line ~114 confirming scope) — medium on real-world
frequency, since no live run in this corpus has yet produced a `carried` resolution (the mechanism
is untested in practice, only in the simulator and by source trace).

---

## Disconfirming pass (checked and held; not re-raised)

- Re-checked whether R2-9's ledger-clause fix (time/access-date trigger) might already be *silently
  implemented* in `debate.js`, which would make the report's "Build now" disposition stale in the
  optimistic direction — direct read of lines 152–156 confirms the ledger clause is still
  prose-change-keyed only; no time/access-date condition present. The report's "Build now" framing
  is accurate (not yet built), not overclaiming. Not a gap.
- Re-checked R2-8's `lanes >= 4` reconciliation for a corresponding code enforcement — none exists
  (the dispatch loop, lines 128–130, has no floor at all, consistent with row 7 being listed as
  still open). The report does not claim this is enforced in code; it is a stated future
  disposition. Not a gap distinct from row 7's existing (already-graded) open status.
- Re-checked whether the CHANGELOG's round-2 R2-9 phrasing ("added the concrete fix... to the same
  ledgerClause string") could be read as claiming a live code edit rather than a report-table
  recommendation. On balance, read in the full row-10 context (disposition explicitly says "Build
  now," future tense), this resolves as "added the concrete *recommended text* to the report," not
  a false completion claim — a plausible ambiguity on a skim, but not misleading on a full read.
  Noted, not raised as a gap.
- Checked the `adjudicated`/`prevGapIds` interaction for a double-counting or gap-resurrection bug
  across non-adjacent rounds beyond what R1's "gap-id rollover" (§2.3 item 3, already known-failing
  and graded) covers — traced through two additional hypothetical round sequences by hand; found no
  behavior beyond what item 3 already names. Not a new gap.
- Checked whether `exhausted`'s computation (line 200) could misfire on the PASS path — confirmed
  it cannot (`redEnv.verdict === 'PASS'` breaks the loop before `exhausted` is evaluated, and the
  guard `redEnv.verdict !== 'PASS'` covers the boundary correctly). Not a gap.

## Verified clean this round (risk-relevant claims re-traced against live, unchanged `debate.js` at `d164ab2`)

Judge dereference still unguarded (line 181/184); `citationPasses` still `const` outside the loop
(line 139/148); ledger clause still prose-change-keyed only (lines 152–156); red-lens bulk /
red-merge judgment routing unchanged (lines 162/167); `lanes = 3` default still has no floor (line
34/128); no `lanes` field in the return object (lines 210–218); no functional `--smoke` parsing
path (line 34's destructure has no `smoke` key). All consistent with the report's existing,
round-2-corrected claims — no new drift, no regression on these points this round.

## Round 3 disposition

FAIL. 3 new gaps, all additive-fix-sized (one guard clause / one `takeFriction` call / one prompt
sentence or one data-plumbing line), none disputing H1–H5 or any prior closure. R3-1 and R3-2 are
the gate tier (a misleading terminal state under a plausible-if-uncommon input; a factual "never
dropped" claim directly falsified by the seat this report's own headline evidence came from).
R3-3 is real but lower-urgency (untested in practice, not yet observed live). Expected convergence:
all three close by the same additive-fix pattern this corpus has used throughout rounds 1–2.
