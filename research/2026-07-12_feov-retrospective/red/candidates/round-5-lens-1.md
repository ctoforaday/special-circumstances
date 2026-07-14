# Round 5, Lens 1 (leaf-node citation verification) — slice: front matter / §0 / §1 (1.1-1.4)

Scope note: per the established 3-instance split for this lens (matching rounds 3-4's lens-1/2/3
division), this pass covers the report's front matter (round-correction summary paragraphs), §0
(Headline), and §1 (1.1-1.4, the doubts). Per the ledger skip-clause, claims in this slice not
touched by `blue/CHANGELOG.md`'s Round 4 entry retain their prior-round confidence and were not
re-fetched; this pass re-verified (a) the new Round-4 summary paragraph (front matter, never
before leaf-audited) and (b) §1.2's R4-5 addition (the only in-slice row the CHANGELOG names as
changed). Live re-check: `git fetch origin main` — HEAD still `42dba2d`, unchanged since round 4's
merge-seat verification; `debate.js` untouched. No drift-based gap in this slice.

Two of this pass's findings surfaced by cross-checking an in-slice claim (the Round-4 summary
paragraph's R4-1 sentence) against its stated evidentiary support, which lives in §2/§3/§4/§5
(lens 2/3 territory). Per protocol ("a change-summary is a navigation hint, never the audit
surface" and the mandate to read the FULL report), both are reported here rather than dropped for
being outside the nominal slice boundary.

## Finding 1 — "four live regression chains" enumerated two different, largely non-overlapping
ways in the same document, and the odd-one-out enumeration is factually wrong on two of its three
non-common entries — MEDIUM

**Location A (front matter, Round 4 corrections summary):** *"this corpus contains four such live
chains, and the judge was never dispatched once across three completed rounds ... independently
corroborated by the project's own backlog, commit `42dba2d`, naming this retrospective's own chain
as its worked example."* This traces to §2.1's gap-id-rollover row, part (b), which gives the
canonical list: *"R1-5 → R2-4 → R3-4/R3-9 ... plus ... R2-5 → R3-10 ... R2-7 → R3-6 ... and
R2-8 → R3-5 → R4-3."*

**Location B (§3, row 23, "Lineage-following contested-gap detection"):** *"not hypothetical: this
exact corpus contains four live regression chains (R1-5→R2-4→R3-4/R3-9; R1-13→R2-1→R3-7;
R1-16→R2-8→R3-5; R2-5→R3-8)..."*

**Problem:** both locations assert "four" regression chains as the evidentiary base for R4-1 (the
docket detector's lineage-blindness), and both cite the same first chain
(R1-5→R2-4→R3-4/R3-9), but the other three entries in each list disagree completely — three
different chain shapes, not a reordering of the same three. Direct leaf-node check against
`red/findings.md`'s own `CLOSED WITH REGRESSION` bookkeeping (grep-confirmed, all 10 instances
read):

- `R1-13 → R2-1` is real (`findings.md` line 743: "R1-13 — CLOSED WITH REGRESSION (round 2) →
  R2-1"). But Location B's `→ R3-7` is not: `findings.md`'s own R2-1 entry states "R2-1 — CLOSED
  (round 3)" (clean, no regression) and explicitly disclaims the link Location B asserts: *"A
  narrower, distinct follow-on — whether occurrence 2 is even the same mechanism — is R3-7, **not
  a reopening**."* Location B cites, as one of its four load-bearing chain examples, a link its
  own corpus's findings file states in plain words is not a chain link.
- `R2-5 → R3-8` in Location B is also wrong per the same file: R2-5's actual regression
  continuation is R3-10 (`findings.md` line 379: "R2-5 — CLOSED WITH REGRESSION (round 3) →
  R3-10"). R3-8 is a *different* finding class on the same footnote — live-source drift on
  `[^CostFigureProvenance]`'s backlog pin (`findings.md` lines 558-570: "live-source drift on the
  report's own most-cited backlog footnote — a textbook self-application of §3 row 10's named
  risk") — not a regression of R2-5's fix.
- `R1-16 → R2-8 → R3-5` in Location B is accurate as far as it goes (all three links confirmed:
  `findings.md` lines 762, 399-403, 97-103) but truncates the chain one link short of its current
  tip — the same chain continues `→ R4-3` (`findings.md` line 97: "R3-5 — CLOSED WITH REGRESSION
  (round 4) → R4-3"), which is a currently-*open* gap this very report addresses at §3 row 6's own
  disposition text a few rows above row 23. Location A's version of this same chain
  (`R2-8 → R3-5 → R4-3`) correctly reaches the live tip but drops the R1-16 origin instead —
  neither location states the full four-link chain.

Net: of the three non-common chain-examples backing R4-1's "four chains, judge never dispatched"
claim, Location A's are individually accurate but under-attributed at the front (missing R1-16),
and Location B substitutes two chain examples that the report's own findings file explicitly
contradicts (R1-13→R2-1→R3-7) or misattributes (R2-5→R3-8), while also truncating the one
correctly-shaped example it shares partial credit for. This is the report auditing a
lineage-tracking defect (R4-1) while itself carrying two different, partly-incorrect lineages for
its own headline supporting example — the report's most load-bearing round-4 finding is backed, in
one of its two stated locations, by mis-traced evidence.

Does not change the underlying verdict on R4-1 itself: the *count* ("four," "zero `### LEAD`
headers," "judge never dispatched") is independently true and re-confirmed this round (HEAD still
`42dba2d`, `debate.js` unchanged, `debate.md` still has zero `### LEAD` headers per round-4's
merge-seat grep). What's wrong is specifically which four examples back it in §3 row 23.

**Corroboration:** HIGH — every sub-claim verified against `red/findings.md`'s own text, not
against blue's paraphrase of it; the two contradicted attributions are direct quotes from
findings.md disclaiming the exact link Location B asserts.

**Required fix:** replace §3 row 23's parenthetical chain list with the same one already correctly
stated at §2.1 (R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3) so the report cites one
consistent evidentiary set for "four chains" in both locations that make the claim — or, if row 23
means to cite genuinely different examples, each substitute must independently trace to a real
`CLOSED WITH REGRESSION` lineage in `findings.md` (which rules out R1-13→R2-1→R3-7 as currently
worded). Cheapest: copy Location A's list into Location B verbatim.

**Grade:** likelihood certain (static text, already present) x impact medium (undermines a
spot-check of the round's single most load-bearing finding's own cited evidence, though not the
finding's core conclusion) x complexity trivial (one clause, copy an already-correct list from
four rows above).

## Finding 2 — cross-corpus PDF-lossiness id list drops one item at one of its three occurrences —
LOW

**Location:** §3, row 13 (PDF extraction) — *"kept 3+ figures at unable-to-corroborate across 4
rounds (memory-architecture corpus's own **MA-R1-19, MA-R1-28, MA-R3-14, MA-R3-15, MA-R4-9** —
prefixed round 4, R4-5)"* — against §2.1's Tier C bullet and §4's rank-1 row, both of which list
the same set **plus** `MA-R2-8 residual`: *"MA-R1-19, MA-R1-28, MA-R2-8 residual, MA-R3-14,
MA-R3-15, MA-R4-9."*

**Problem:** three locations in this report enumerate the same cross-corpus id set (all corrected
to the `MA-` prefix this round, per R4-5); two list six ids, one lists five, omitting
`MA-R2-8 residual`. Direct check against the memory-architecture corpus's own findings file
confirms `MA-R2-8`'s "residual" phrasing is real and PDF-extraction-relevant — its own "Friction"
section states verbatim: *"A full-PDF-text-search / PDF-table-extraction tool would discharge
R1-19, R1-28, R2-8's residual, and R2-10 definitively"* (`research/2026-07-12_memory-architecture/
red/findings.md`, line 688) — so the six-item version (§2.1, §4) is the more complete one and
row 13's five-item version is the incomplete outlier, not the reverse.

**Corroboration:** HIGH (direct grep of all three locations; direct read of the memory-architecture
source line naming R2-8's residual as PDF-extraction-blocked, same status as R1-19/R1-28).

**Required fix:** add `MA-R2-8 residual` to row 13's list to match §2.1/§4.

**Grade:** likelihood certain x impact low (a redundant enumeration losing one of six items;
doesn't affect the "3+ figures" headline count, which is already a floor, not the omitted item's
own count) x complexity trivial (one insertion).

## Checked and held (this slice)

- Front matter's R4-2/R4-3/R4-4/R4-5 summary sentences: each compared verbatim against its cited
  row/footnote (§3 row 20 + §2.3 addition 13; §3 row 6; risk-accepted closing paragraph; §1.2 +
  [^GapIdScheme]) — all match. HIGH.
- Report-wide grep `"4th|fourth"`: zero remaining uncorrected stale-numeral instances; the three
  matches present are the R4-4 correction note itself, an unrelated "fourth round" mention, and the
  already-corrected row 15 prose. R4-4 closure holds. HIGH.
- §1.2's R4-5 addition (`MA-R2-10` naming-discipline paragraph) verified against
  `research/2026-07-12_memory-architecture/red/findings.md` lines 423-427 (the actual MA-R2-10 gap:
  `[^SingleUserLowRisk]` citing "practitioner consensus" for blue's own cross-lane synthesis) —
  exact match, distinct from a lane-provenance failure as the report itself argues. HIGH.
- `[^GapIdScheme]`'s "memory-architecture ... confirmed live ... to run at least to R4-12" —
  re-confirmed live: memory-architecture `red/findings.md` running gap ids do reach R4-12 (12
  distinct round-4 gaps, `### R4-1` through `### R4-12` all present). HIGH.
- Red-auditor memory-pattern-file count: live count now 25 (was 23 at round 4). Consistent with
  the already-established round-2/3/4 disposition (expected growth of a live accreting mechanism,
  not a gap) — not re-raised, per precedent against re-litigating a settled non-gap every round.
- `main` HEAD: re-fetched, still `42dba2d`; `debate.js` unchanged since round 4's merge-seat
  verification. No new drift in this slice.

## Ledger additions

See `red/citation-ledger.md` — this pass's verified claim/reference/confidence lines appended
under "round 5".
