# Round 5, lens 3 (leaf-node citation verification) — slice: §4 / §5 / Footnotes

Scope discipline: sections divided evenly among 3 instances; this pass audits `## 4`, `## 5`, and
`## Footnotes` of `blue/report.md`. Full living report re-read in context (all 911 lines), not just
the CHANGELOG diff.

**Pre-check:** `blue/CHANGELOG.md` has no "Round 5" entry — blue has not submitted new changes since
Round 4. `git log --oneline -5 origin/main` still HEAD `42dba2d` (no drift since the round-4 merge-seat
check); `git diff 42dba2d origin/main --stat` empty. Per the ledger skip-clause, all round ≤4
HIGH-confidence verifications in this slice stand and were not re-fetched. This pass instead re-derives
several claims from their primary source that the ledger had *not* yet leaf-verified for this specific
angle (staleness of a claimed-still-open status), which the skip-clause does not cover (it exempts
re-fetching already-verified statements, not auditing a new angle on an old citation).

---

## New finding (R5-1): §4 row 1's "blocks memory-architecture's own MA-... from resolving" list cites four already-CLOSED ids as if still open, and two of the six are not "unable-to-corroborate" cases at all

**Location:** §4, row 1 (rank 1, "PDF full-text / table extraction") — *"**Open — highest-value
unbuilt capability**; blocks memory-architecture's own **MA-R1-19, MA-R1-28, MA-R2-8 residual,
MA-R3-14, MA-R3-15, MA-R4-9** (prefixed round 4, R4-5 — see [^GapIdScheme]) from resolving past
'unable-to-corroborate-at-leaf-node'..."*

**Problem:** direct read of `research/2026-07-12_memory-architecture/red/findings.md` (the primary
source for every one of these ids) shows the current status of each:

| id | current status (per that corpus's own record) | nature |
|---|---|---|
| R1-19 | **OPEN** — line 118 self-summary: "Carried: R1-19 (agent-PR figures, friction-blocked)" | genuine PDF/lossy-fetch block |
| R1-28 | **CLOSED, round 3** — "All three compounding sub-defects discharged this round... Red accepts closure." | closed via live abstract re-fetch + re-citation, not PDF extraction |
| R2-8 residual | **CLOSED, round 3** — "Contradicted number gone — red accepts closure." (both legs re-verified live) | was a contradicted-figure miscitation, not a lossy-fetch case; resolved without PDF access |
| R3-14 | **CLOSED, round 4** — "CLOSED, but the re-homing spawned R4-10... The over-attribution red flagged is gone." | closed by claim-scope trim, not PDF extraction |
| R3-15 | **CLOSED, round 4** — "CLOSED. `[^RecMem]`... now read 'up to ~87%...' Re-verified live at the leaf node this round." | was an over-claim beyond what the abstract said (abstract stated a *stronger* result); resolved via abstract re-fetch, not blocked |
| R4-9 | open (fresh round-4 finding, no closure recorded — corpus has no round 5) | but per its own text, verified via "three independent routes (abstract fetch, full-text HTML, web-search)" — a diagnosed miscitation, not an "unable-to-corroborate" block |

So of the six ids cited as evidence the PDF-extraction gap is currently blocking resolution, **four
(R1-28, R2-8, R3-14, R3-15) are closed**, and of the two still open, **one (R4-9) is a fully-diagnosed
miscitation, not a case blocked by lossy PDF fetch** — only R1-19 is a genuine, presently-open,
PDF-lossy-fetch-blocked instance. The report's own §2.1 Tier C bullet (line 517, same six-id list) and
§3 row 13 (line 715, a *five*-id list that drops R2-8 entirely) disagree with each other and with this
row on membership — three near-identical enumerations of "the memory-architecture corpus's PDF-blocked
gaps" inside the same document, no two of which match, and none of which reflects the corpus's own
current closure record.

The underlying source of "R2-8's residual" traces to `red/findings.md`'s own "Friction (carried
forward, unresolved)" section (line 688): *"A full-PDF-text-search / PDF-table-extraction tool would
discharge R1-19, R1-28, R2-8's residual, and R2-10 definitively."* — written mid-round-2, before R1-28
and R2-8 were closed the following round (round 3) by ordinary live re-fetch, not by a PDF tool. That
source sentence's own prediction was falsified by what actually closed those gaps — a second-order
staleness the FEOV report inherited without checking, on top of the first-order staleness (citing
`R2-8 residual`/`R1-28` as open at all).

**Why this matters, graded:** the row's overall disposition (build PDF extraction, it's the #1 tool
gap) is not undermined — it is independently corroborated by the live backlog (`ideas/backlog.md`
line 27(c), re-checked this round, still verbatim "TOP TOOL GAP... PDF full-text/table extraction")
and by R1-19/R4-9 alone. But the specific evidentiary claim as written — "blocks six named gaps from
resolving" — overstates the current blocking count by roughly 2x (2 genuinely blocked vs. 6 cited),
and mischaracterizes the *mechanism* by which four of the six were actually resolved (ordinary
re-fetch, which the row's framing implies cannot happen without the tool). A reader relying on this
row's citation list to judge how much value the PDF tool would unlock is misled toward a larger
number than the source supports.

**Corroboration confidence:** LOW for the row's "blocks... MA-R1-28, MA-R2-8 residual, MA-R3-14,
MA-R3-15" sub-claims as currently phrased (source shows all four closed); MEDIUM for MA-R4-9 (real,
open, but wrong failure class — miscitation, not lossy-fetch); HIGH for MA-R1-19 (matches source
exactly) and for the row's #1-ranking disposition itself (independently corroborated by the live
backlog).

**Grade:** likelihood of reader-misdirection — high (the sentence is read at face value, no hedge);
impact — medium (does not flip the build-it disposition, which survives on independent grounds, but
inflates the stated evidence for it and the three-way internal inconsistency invites a "which list is
right" question a stickler should not have to resolve); complexity-to-fix — low (either drop the four
closed ids and cite only R1-19 + a general "this class recurs" framing, or keep the historical list
but add "as of memory-architecture round 2; four of six since closed by ordinary re-fetch, not by a
PDF tool" — and reconcile the three locations' membership to one canonical list). **Pattern:
citation-status drift / closed-not-flagged, recurring across three locations in the same document.**

---

## Re-confirmed, no change (high confidence, not re-fetched per skip-clause — listed for completeness of this slice's coverage)

- §4 row 1 role/round attribution ("3/3 roles — red r1–r4, blue r1–r4, judge r2") — ledger round 1,
  HIGH.
- §4 row 2 CVE/advisory role/round attribution and engineered-around disposition — ledger round 1,
  HIGH.
- §4 rows 3–10 (undefined-path merge status, write-block role count, ENAMETOOLONG 2/2-not-3/3 count,
  live-source-drift single-instance, long-tail singletons) — ledger rounds 1–3, HIGH; re-read in full
  this pass, text matches ledger's last-verified state, no CHANGELOG touch this round.
- §5 items 1–12, including new item 12 (R4-1 independence statement) — ledger round 4 merge-seat,
  HIGH; re-read in full, unchanged since.
- Footnote [^GapIdScheme] (R4-5) — re-verified this pass directly against
  `research/2026-07-12_memory-architecture/red/findings.md` lines 118 and 643 (R4-12 present, corpus
  runs to at least R4-12 as claimed) — HIGH, confirms ledger.
- [^CostFigureProvenance]'s re-pin at `d164ab2` — unchanged, no drift since round 3/4 verification.
- Report-wide "4th|fourth" grep (R4-4) — not re-run this pass (already closed at round 4 merge-seat,
  no section touched); accepted per skip-clause.

No other new leaf-node discrepancies found in this slice this round.
