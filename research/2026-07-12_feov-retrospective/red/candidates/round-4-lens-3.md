# Round 4, Lens 3 of 3 — leaf-node citation verification (slice: §4 Friction ranking + §5 Open
questions + Footnotes)

Full living `blue/report.md` re-read in context this pass (all 833 lines), not just
`blue/CHANGELOG.md`'s diff. Per convention set in round 3 (lens 1 = §0/§1, lens 2 = §2/§3, lens 3 =
§4/§5/Footnotes), this pass covers §4, §5, and the Footnotes block. Ledger skip-clause honored:
claims already HIGH in `red/citation-ledger.md` whose section was not touched by the most recent
`blue/CHANGELOG.md` entry (Round 3, 2026-07-14) were not re-fetched. Re-verified fresh: the
Round-3-touched content that falls in this slice — §4 row 5 (R3-7) and footnote
`[^CostFigureProvenance]` (R3-8) — plus a full live re-fetch of `origin/main` and the retrospective's
own input corpus, since this slice's footnotes are the report's highest-drift-risk citations.

## Live re-verification this round

`git fetch origin main && git log --oneline -5 origin/main`: HEAD has advanced **again**, past the
report's newest pin. `d164ab2` (the R3-8 pin) → **`42dba2d`** (2026-07-14T00:49:14-07:00), one more
docs-only commit to `ideas/backlog.md`. `git diff d164ab2 42dba2d -- .../debate.js`: empty — no
code drift, consistent with every prior round's pattern. `git diff d164ab2 42dba2d --stat`:
`ideas/backlog.md | 2 +-` (1 line changed). `git grep -ni "independen"
plugins/frank-exchange-of-views`: still **zero** hits at `42dba2d` — R3-6's fix holds.
`inputs/run2-friction.md` line 4 / `blue/CHANGELOG.md` Round 0: re-read directly, textual
distinction from R3-7 (length-ceiling vs. "shell parsing") still holds exactly as stated — **R3-7
confirmed still accurate, no regression.** `ideas/backlog.md` item 28 (cost audit, R3-8's subject):
byte-for-byte unchanged between `d164ab2` and `42dba2d` — the diff touched a *different* backlog
item entirely (see below). §4 row 5 and footnote `[^CostFigureProvenance]` both verified clean,
content-wise; only the SHA pin is one commit stale (harmless per the established
docs-only-drift precedent — not re-raised as its own gap, consistent with round-2/3's norm of
only flagging drift where materially load-bearing).

## New gap (round 4)

**R4-1 — HIGH — certain (already realized 3+ times in this corpus, confirmed by direct code trace)
x high (the report's own praised "cannot be fooled" adjudication mechanism has never once engaged
for a defect recurring across rounds) x low-medium (a stated fix already exists as a live external
proposal) — corroboration: HIGH (both halves — the code mechanism and the self-referential
instance — independently verified by direct read, not inference).**

**Location:** §5 (Open questions) has no item for this; the closest existing text is §2.1's Tier A
table, *"Gap-id rollover across non-adjacent rounds: `prevGapIds` holds only the prior round, so a
gap closed in round 1 recurring in round 3 classifies 'new,' not 'contested'"* (line ~458, outside
this instance's §4/§5/Footnotes slice, flagged here per the round-3 precedent of raising
out-of-slice findings discovered while auditing in-slice citations — this one surfaced while
re-fetching `ideas/backlog.md` for the in-slice `[^CostFigureProvenance]` footnote check above).

**Problem:** a new backlog commit, `42dba2d` (2026-07-14T00:49:14-07:00, "docs(backlog): docket
detector tracks IDs not lineages — regression-chain gaps evade the judge"), documents a *distinct*
and *more severe* variant of the gap-id-rollover row that the report does not cover anywhere. Direct
verification against `debate.js` at `42dba2d`:

```
const contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))
...
prevGapIds = gapIds
```

`contested` is computed by literal id membership only. The existing §2.1 row concerns *the same id*
reappearing after skipping a round (fixed, per that row, by widening `prevGapIds` to full
adjudicated history). The backlog's finding is orthogonal and unfixed by that proposed remedy: when
red closes a gap "WITH REGRESSION" and mints a **successor gap under a fresh id** for the new,
smaller defect the fix introduced — which is exactly this retrospective's own documented practice,
e.g. R1-5 (closed) → R2-4 (successor id) → R3-4/R3-9 (further successor ids), all for what is
substantively one recurring defect in the same footnote — `prevGapIds.has(g.id)` **never matches**,
no matter how much history is retained, because the id itself changes every hop. `contested` stays
empty, the judge branch (`if (contested.length > 0)`) never fires, and (per the loop trace already
independently confirmed for R3-1/R3-2) the only remaining brake on an indefinitely-recurring,
ever-renamed defect is the `maxRounds` cost ceiling — not adjudication. The backlog commit names
this exact chain (`R1-5 → R2-4 → R3-4/R3-9`) as its own worked example, i.e., the project's own
live tracking already recognizes this retrospective's `[^DiminishingReturns]` saga (independently
flagged by the report itself, footnote text: *"the third round running a defect recurred in this
one footnote (R1-5, R2-4, now R3-9)"*) as the demonstrating instance — the report names the
symptom (a footnote keeps breaking) but has never connected it to the *mechanism* consequence (this
exact multi-round recurrence was never once eligible for judge review, despite §0's claim that the
regression suite/adjudication path "cannot be fooled by a stale status line").

**Required fix:** add a graded row (§3, out-of-slice) crediting the backlog's proposed remedy —
`supersedes: [prior-ids]` on the gap envelope; contested-detection follows supersession chains
(chain depth ≥ 2 → docket) rather than raw id membership; a lineage regression-test case for the
simulator (§2.3, out-of-slice) — and a new §5 open question: *does widening the existing
"gap-id rollover" fix (§2.1's proposed remedy: extend `prevGapIds` to full history) also require
lineage-following, or are these two independent code changes?* (They are independent per the trace
above — same-id-after-gap and different-id-successor-chain are different failure modes needing
different fixes; the report should not let one fix's rationale imply it covers the other.) Cheap to
state; the actual code fix is appropriately docketed for run 4, consistent with how R3-1/R3-2's
code-side fixes are already handled (report-side correction now, code fix later — same disposition
should apply here).

## Verified clean this round (§4/§5/Footnotes; unchanged since last touched, per skip-clause)

- §4 rows 1–4, 6–10 and the shape-verdict paragraph: no citation in this slice was touched by
  Round 3's CHANGELOG; spot-re-read against `run2-friction.md`/`run1-friction.md`/`blue/frontier.md`
  confirms no drift in the underlying archived files (these are frozen historical inputs, not live
  repo state, so drift risk is structurally near-zero for them).
- §5 items 1–11: internally consistent with their cited resolutions (R2-3, R2-6, R2-7, R2-10 as
  reflected); no footnote citation in §5 itself needs re-verification this round (§5 was not
  touched by Round 3's CHANGELOG).
- Footnotes: `[^Run2Friction]`/`[^FrictionCount]` (21 entries, re-counted directly again this round:
  unchanged), `[^HookGrep]`/`[^NoPackageJson]`/`[^GoTests]` (local, static, re-spot-checked),
  `[^PdfMcp]` (not touched this round, not re-fetched per skip-clause), `[^SubagentWriteBug]`,
  `[^WisdomCrowds]`, `[^AgentTestTiers]`, `[^ProvenanceSurvey]`, `[^ClaimManifest]` — all
  untouched by Round 3, held at their prior HIGH/MEDIUM grades per ledger, not re-fetched.
- `[^DiminishingReturns]` (touched by R3-9, but that edit lives inside §1.1's territory, lens 1's
  slice): the footnote *definition* physically sits in my Footnotes slice, so spot-checked here too
  — the sense-1/sense-2 disambiguation reads as internally coherent on direct re-read, matches the
  independently-reconfirmed arXiv:2606.02646 figures already in the ledger (GSM-Hard, knee≈10,
  ~1.8-agent plateau by N=30). No new issue found in the footnote text itself; R4-1 above is a
  structural/mechanism finding sparked by this footnote's own recurrence history, not a citation
  defect in the footnote.

## Disposition

FAIL (this slice, this round). One new gap (R4-1, HIGH), sparked by a live external source
(the project's own backlog, freshly committed at the moment of this audit) that the report has not
yet had a chance to incorporate — this is exactly the class of finding leaf-node citation
verification exists to catch: the primary source moved and now says something new and directly
relevant. Everything else in §4/§5/Footnotes verified clean or correctly unchanged. R4-1's real
home is §2.1/§3 (lens 2's territory) and §5 (mine); flagging in full here per the stickler mandate
rather than assuming lens 2 will independently find it — the merge seat should dedupe/anchor
against lens 2's pass if it also caught the same backlog commit.
