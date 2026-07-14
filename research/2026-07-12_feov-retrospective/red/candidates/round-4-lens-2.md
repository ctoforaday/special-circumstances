# Red round-4 lens 2 — leaf-node citation verification, slice 2 of 3 (§2 Testing strategy, §3 What should change)

Scope: this instance took §2 (lines 445–627) and §3 (lines 628–676) of `blue/report.md`, per the
even three-way split of the report's section headings (instance 1: §0–§1; instance 2 [this file]:
§2–§3; instance 3: §4–§5–Footnotes). Read the full living report in context (not the diff), the
full `blue/CHANGELOG.md`, and the full `red/findings.md` before auditing. Consulted
`red/citation-ledger.md` and applied the skip-rule: claims verified HIGH in a prior round and
untouched by this round's CHANGELOG entries were not re-fetched. Re-fetched `origin/main` live
(the repo is present locally at `C:/Users/gbloc/Projects/special-circumstances`) and read
`debate.js` directly rather than trusting prior-round quotes, because HEAD has moved again.

## Live-repo state at audit time

`git fetch origin main` → HEAD is `42dba2d`, one commit past the report's last-pinned `d164ab2`.
`git diff d164ab2 42dba2d --stat` → `ideas/backlog.md` only (1 line changed), `debate.js`
byte-identical. No report claim about `debate.js` is invalidated by this drift. But the new
backlog line itself is a load-bearing live finding — see Finding 1.

## Finding 1 (NEW, round 4) — the contested-docket detector is lineage-blind by construction; this retrospective's own gap chain is the live proof, and it now appears in the live backlog uncited by the report

**Location:** §2.1, Tier A table, row *"Round loop / contested docket / deadlock / safety ceiling
/ `adjudicated` bookkeeping | `workflow.js` lines 113–166 | Pure `Set`/array logic over canned
envelope shapes; currently zero tests on `main` [all lanes]"* (plus its R3-1 addendum, which covers
only the schema-legal-empty-`gaps` degenerate shape) — and the adjacent, already-known row
*"**Gap-id rollover across non-adjacent rounds**: `prevGapIds` holds only the prior round, so a
gap closed in round 1 recurring in round 3 classifies 'new,' not 'contested'"* (§2.1) / *"Gap-id
rollover — id present r1, absent r2, present r3... known-failing until `prevGapIds` widens to full
adjudicated history"* (§2.3 addition 3).

**Problem:** the existing rollover row assumes the *same literal gap id* reappears after skipping
a round. That is not how this retrospective's own gap-numbering actually works, and the report
never covers the mechanism that is actually in play. Direct read of the live `debate.js` (fetched
this round, `origin/main` @ `42dba2d`, unchanged since `47ae48d`):

```
const gapIds = new Set(redEnv.gaps.map(g => g.id))
const contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))
const hasNew = [...gapIds].some(id => !prevGapIds.has(id))
if (contested.length > 0) { /* ...dispatch judge... */ }
...
prevGapIds = gapIds
```

`contested` is computed by exact string-equality on `g.id` against the *immediately prior* round's
id set. `RED_ENVELOPE`'s gap schema (read directly, same file) has no `supersedes`/lineage field —
`id` is a bare string with no required relationship to any prior gap. This retrospective's own
findings.md is the demonstration: the same underlying defect (an uncited precise figure in
`[^DiminishingReturns]`) was raised as **R1-5**, closed-with-regression, re-raised as **R2-4**,
closed-with-regression (twice: as **R3-4** and **R3-9**) — three different id strings across three
rounds for what the report's own footnote text calls "the third consecutive round of trouble in
this exact footnote." Under the live code, `contested` never includes any of R2-4, R3-4, or R3-9,
because none of their id strings was ever a member of `prevGapIds` — a freshly-numbered
regression-closure successor is indistinguishable from a genuinely new, unrelated gap to the loop.
The judge is therefore never dispatched for this exact, live, repeatedly-observed pattern, no
matter how many rounds it persists; **the only backstop is the `maxRounds` cost ceiling.** This is
a materially different (and more severe) failure than the already-documented rollover row: rollover
assumes ids are stable and only breaks on a skipped round; this breaks on *every* round, adjacent
or not, for the report's own dominant recurring-defect class (repair-regression), because blue's
practice (and the whole additive-fix protocol) is to mint a new id for each round's fix.

**Why this is new, not a restatement of an audited item:** confirmed by grep against
`red/findings.md` and the "Noted, not raised" sections of rounds 2–3 — no prior round's gap
(R1-11, R2-*, R3-1..R3-10) names id-vs-lineage tracking; §2.3 addition 3 explicitly frames the
known gap as "id present r1, absent r2, present r3" (same-id, non-adjacent), not "different id
every round" (any-adjacency). Also confirmed live: `ideas/backlog.md` gained a new item at `main` @
`42dba2d` (one commit past the report's `d164ab2` pin, fetched this round) titled *"frank-exchange-
of-views: the docket detector tracks IDs, not lineages"* — written from the operator's own
observation of run 3 (this retrospective), independently converging on the same code-level cause
I traced from `debate.js` directly (not merely trusted from the backlog): *"red closes gaps 'WITH
REGRESSION' and mints successor gaps under fresh IDs (R1-5 → R2-4 → R3-4/R3-9)... the contested
docket never arms, and the judge never sees a dispute lineage no matter how long it persists — the
only remaining brake is the maxRounds cost ceiling."* The backlog item also proposes the fix: a
`supersedes: [prior-ids]` field on the gap envelope, lineage-aware contested-detection (depth >= 2
→ docket), and a simulator lineage-regression test.

**Corroboration confidence: HIGH.** Verified two ways independently: (a) direct read of the live
`debate.js` source at the current HEAD (not quoted from a prior round or from the backlog), showing
the exact `id`-only equality check and the schema's lack of a lineage field; (b) this retrospective's
own findings.md id sequence (R1-5/R2-4/R3-4/R3-9), which is primary-source proof the mechanism has
already fired, not a hypothetical. The backlog entry is a third, convergent, but not load-bearing,
corroboration (an operator self-report I did not need to rely on).

**Grading:** likelihood HIGH (already realized three times in this exact corpus, not speculative) ×
impact HIGH (the contested-docket/judge-adjudication step is the only structural convergence check
this system has for exactly the failure mode — repeated, regressing, but never-quite-closed fixes —
that both this retrospective's own findings and its own disposition text ["severity is declining
monotonically... convergent, not divergent"] currently treat as an empirical trend rather than a
mechanism-backed guarantee) × complexity LOW-MEDIUM (schema field + one filter-logic change +
one simulator case, per the backlog's own already-drafted fix — no new infrastructure).

**Required fix:** add the `supersedes` field to the gap schema and thread lineage into the
contested-docket filter (the code fix may be docketed for run 4, consistent with this report's own
practice for R3-1/R3-2); report-side, correct §2.1's rollover row and §2.3 addition 3 to state the
two distinct failure classes side by side (same-id-skips-a-round vs. new-id-every-regression-round),
and add a row to §3's graded table naming this fix explicitly — it is currently invisible to the
"what should change before run 4" table entirely.

## Finding 2 (NEW, round 4) — R3-5's arithmetic fix resolves the total by fiat without editing the ambiguous sentence that caused the miscount

**Location:** §3 row 6 — the still-unedited R1-16 disposition clause *"assign the
critical-stance/adversarial-disconfirming lens to at least 2 of N lanes (not 1-of-N)"* — versus
the round-3 correction appended to the same row: *"four named methods (primary-literature /
practitioner-production / adversarial-disconfirming-first / local-repo critical-stance), **one of
which (adversarial-disconfirming-first) carries a 2-of-N redundancy floor** per item 6's own text
above."*

**Problem:** R3-5 (closed this round per findings.md) fixed the *arithmetic* (3 unfloored × 1 lane
+ 1 floored × 2 lanes = 5, not 4) by asserting a specific reading of the original, still-unedited
sentence: that "the critical-stance/adversarial-disconfirming lens" names only one of the four
items (adversarial-disconfirming-first) as floored, and that "local-repo critical-stance" is a
separate, unfloored fourth method. But the original sentence itself is not edited to say that, and
its own punctuation reads the opposite way on a plain parse: "the critical-stance/adversarial-
disconfirming lens" (singular "lens," slash joining two adjectives into one compound name) is the
exact same construction the four-method roster uses to *separate* distinct items ("primary-literature
/ practitioner-production / ..."). The same slash character is used two ways in the same row —
as a synonym-joiner in one sentence, as a list-separator two sentences later — with nothing in the
text flagging the switch. A reader who stops at the (still-present, unedited) floor sentence will
reasonably conclude the *combined* critical-stance/adversarial-disconfirming pair is the floored
unit (yielding a 3-not-4-method roster, and total lanes = 1+1+2 = 4 — the very total R3-5 just
declared wrong). R3-5's own required fix offered two options: "(a) state the true minimum as
`lanes >= 5` with all four methods distinct, or (b) state explicitly that the two adversarial/
critical methods collapse into one floored lens... and why that merge is safe." What shipped is a
third thing: option (a)'s conclusion, asserted in new prose appended after the original sentence,
without ever touching the original sentence that reads like option (b). The correction argues past
the ambiguity rather than removing it.

**Corroboration confidence: HIGH** — pure textual comparison within the shipped document, no
external source needed; both sentences quoted verbatim above from the current `blue/report.md`.

**Grading:** likelihood HIGH (a reader encountering the floor-assignment sentence in its normal
reading-order position, before the later corrective paragraph, picks up the wrong count) × impact
LOW-MEDIUM (the correct arithmetic is stated later in the same row, so a full read resolves it;
damage is confined to a reader who stops early, or who greps only the disposition's original
sentence) × complexity TRIVIAL (reword the original floor sentence to name the two items
separately, e.g. "assign the local-repo critical-stance lens to 1 of N lanes and the
adversarial-disconfirming-first lens to at least 2 of N lanes" — one clause, no new research).

**Required fix:** edit the original R1-16 sentence itself (not just append more prose after it) to
remove the slash-as-synonym reading, consistent with R3-5's own already-argued four-distinct-methods
conclusion.

## Finding 3 (NEW, round 4) — the round-2-corrected "third occurrence" trigger for ENAMETOOLONG regressed back to "4th" in the risk-accepted summary line, a fifth location the R2-1 propagation missed

**Location:** §3, the "Explicitly risk-accepted" paragraph closing the section — *"ENAMETOOLONG
tooling (#15, re-graded round 1 to track recurrence but still risk-accepted pending a 4th
occurrence) [L1]"*.

**Problem:** R2-1 (closed round 3, per `red/findings.md`) corrected the ENAMETOOLONG build-trigger
count from "fourth" to "third" everywhere it was checked: row 15's own disposition (*"build the
chunking helper if it recurs a third time (not a fourth — corrected trigger, R2-1)"*), §4 row 5
(*"tracked for a third occurrence, not a fourth"*), and §5 item 10 (*"the trigger is a third, not a
fourth"*). `grep -n "4th|fourth" blue/report.md` (this round) turns up exactly one uncorrected
instance: this end-of-§3 summary sentence, which still reads "pending a 4th occurrence" with no
"(not a fourth)" annotation and no cross-reference to the correction. Same class as R3-10 (a
repaired figure/count reaching some reading-order locations and not others) — this is the R2-1
count's own unpropagated fifth location, caught this round because R2-1 was closed as clean without
a report-wide grep for the retracted numeral, only for the three locations the original gap named.

**Corroboration confidence: HIGH** — direct grep + read of the current document; the correction
target (row 15's own sentence, three paragraphs earlier in the same section) is present verbatim in
the same file for direct comparison.

**Grading:** likelihood CERTAIN (static text, not probabilistic) × impact LOW (a reader of only the
risk-accepted summary — plausible, since it is written as a scanning aid — walks away with the
stale trigger count; the correct count is one paragraph away in the same section) × complexity
TRIVIAL (change "4th" to "third," matching row 15's phrasing).

**Required fix:** edit the risk-accepted-list clause to read "...pending a third occurrence
(corrected R2-1)..." — one word, matching the three other already-corrected instances.

## Verified clean this round (no new gap; added to `red/citation-ledger.md`)

- `debate.js`'s contested-docket loop (`prevGapIds`/`gapIds`/`contested`/judge-dispatch, lines
  ~142, 176–187) and `RED_ENVELOPE`'s gap schema (no lineage field) — direct read at `origin/main`
  @ `42dba2d`, this round. HIGH.
- `git diff d164ab2 42dba2d --stat` — `ideas/backlog.md` only, `debate.js` untouched. HIGH (drift
  confirmed harmless to all standing §2/§3 code-level claims except as noted in Finding 1).
- §2.1 Tier A round-loop row's R3-1 addendum (schema-legal `{verdict:'FAIL', gaps:[]}` guard
  description) — re-read against the same live `debate.js`; internally consistent with the code
  (no `PASS`-only break path is affected by an empty-`gaps`/`FAIL` envelope; `contested` computed
  over an empty array is empty; loop falls through to re-dispatch). No regression found.
- §2.1 Tier A friction-aggregation row's R3-2 correction (`takeFriction` called at exactly
  `red-merge`, `judge`, `blue-respond`; never `blue-synthesize`) — re-confirmed against the same
  live file this round (call sites at the lines the report cites). HIGH, unchanged.
- §3 rows 20/21/22 (the three new round-3 rows for R3-1/R3-2/R3-3) — read against the live code;
  descriptions and required fixes remain accurate and undisturbed by the `42dba2d` drift.
- §3 row 19's "zero hits anywhere, including inside the ledger clause itself" (R3-6 fix) — not
  re-run this round (no CHANGELOG-flagged change to this section, ledger already HIGH from round 3
  merge-seat re-grep); reused per the skip-rule.
- §3 row 18 / [^CostFigureProvenance]'s R3-8 re-pin (`d164ab2`, sub-item (d)) — footnote content
  read directly this round; still accurate as of `42dba2d` (the new commit only added the
  docket-lineage backlog item, not a further edit to item 28). No further drift to report.
- §2.3's 11-test enumeration and 14 additions (1–14) — spot-read against the item list for internal
  numbering consistency (no gaps, no duplicate numbers, additions 13/14 correctly marked
  KNOWN-FAILING pending their respective §3 rows). Clean.
- Row 13's PDF-extraction tool claims ([^PdfMcp]) — untouched by round-3 CHANGELOG in this section;
  reused at HIGH per the skip-rule (last verified live, round 2 merge).

## Summary disposition (this lens only)

3 new gaps in §2–§3 for round 4: one HIGH-severity structural finding (docket lineage-blindness,
independently code-verified, not merely a citation defect) that the report currently has no row
for at all; two LOW/LOW-MEDIUM repair-propagation misses (an unresolved naming ambiguity R3-5's fix
argued around instead of editing; a stale "4th occurrence" numeral in one summary sentence R2-1's
propagation missed). No prior round-1/2/3 gap in this slice is reopened; all round-3 fixes touching
§2/§3 (R3-1, R3-2, R3-3, R3-5, R3-6, R3-8, R3-10) were re-checked against the live, once-more-drifted
`main` and hold clean except as noted above.
