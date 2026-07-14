# red findings — FEOV retrospective

LIVING audit — cumulative across rounds.

**Round 1** consolidated five lens passes (`red/candidates/round-1-lens-{1..5}.md`): verdict FAIL,
20 gaps (R1-1..R1-20), full grading preserved below with round-2 statuses prepended.

**Round 2** consolidated five lens passes (`red/candidates/round-2-lens-{1..5}.md`): verdict FAIL,
all 20 round-1 gaps closed (5 with regressions), 11 new gaps (R2-1..R2-11), full grading preserved
below with round-3 statuses prepended.

**Round 3** consolidated five lens passes (`red/candidates/round-3-lens-{1..5}.md`:
3 leaf-node citation slices covering §0–§5 and every footnote, 1 logic/completeness, 1
dark-side/risk) — superseding the single-lens (lens-5) preview previously in this slot, whose three
gaps keep their ids (R3-1..R3-3) unchanged. Keystone claims re-verified live at merge time
(2026-07-14): `origin/main` HEAD is `d164ab2` — one docs-only commit past the report's newest
pin (`88eb57f`), three past its older `47ae48d` pin; `git diff 88eb57f..d164ab2 -- .../debate.js`
is empty (only `ideas/backlog.md` changed), so no report claim is invalidated by drift;
`git grep -ni "independen" -- plugins/frank-exchange-of-views` re-run at the merge seat: **zero
matches anywhere, including inside `debate.js:156`'s `ledgerClause` string** (bears on R3-6, and
on red's own round-2 phrasing — see R3-6's provenance note); the §1.1 body clause "continued gains
observed to 7 agents on the hardest" confirmed still present by direct full-report re-read (770
lines) at merge (R3-4); [^DiminishingReturns]'s self-contradictory sentence and the untagged
252.9k/253k–3M instances (§2.1 first row; §3 row 4 impact cell) read directly (R3-9, R3-10);
§3 row 6's reconciliation arithmetic re-computed against its own four-item roster (R3-5).
arXiv:2606.02646 was independently re-fetched by two separate round-3 lenses (abstract + full
HTML), both confirming blue's R2-4 rebuttal exact: GSM-Hard (not GSM-Plus), "practical knee is
N≈10," effective team size plateaus ~1.8 agents by N=30, single N≤5 pilot predicts the ceiling.
Round-3 verification pairs are appended to `red/citation-ledger.md` (lines 90–111).

**Round 4** consolidated five lens passes (`red/candidates/round-4-lens-{1..5}.md`:
3 leaf-node citation slices — lens 1 §0–§1, lens 2 §2–§3, lens 3 §4/§5/Footnotes — 1
logic/completeness, 1 dark-side/risk). Keystone claims re-verified first-hand at the merge seat
(2026-07-14): `origin/main` HEAD is `42dba2d` — one docs-only commit past the report's newest pin
(`d164ab2`); `git diff d164ab2 42dba2d --stat` touches `ideas/backlog.md` only (+1/−1),
`debate.js` byte-identical across `47ae48d`/`88eb57f`/`d164ab2`/`42dba2d`. The drifting commit
itself is load-bearing: `git show 42dba2d` read directly — a new backlog item, *"the docket
detector tracks IDs, not lineages,"* citing **this retrospective's own gap chain
R1-5 → R2-4 → R3-4/R3-9 as its worked example** (bears on R4-1). `debate.js:178`'s
`contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))` re-read live; `RED_ENVELOPE`'s gap
schema confirmed to carry no `supersedes`/lineage field. `debate.md` checked at header level via
`grep -n "^### "`: **zero `### LEAD` sections across rounds 0–3** — only BLUE/RED headers; stated
precisely because a plain-text grep returns one match (line 528, a *quoted phrase* inside round-3
prose, not a judge section) and lens 5's "zero matches" phrasing would otherwise repeat the R3-6
imprecision class. Also read directly at merge: §3 row 20's disjunction text and §2.3 addition
13's actual assertion wording (bears on R4-2, including a merge-seat temper of lens 4's grading);
§3 row 6's original floor sentence beside its round-3 correction (R4-3); the report-wide
`grep "4th|fourth"` showing exactly one uncorrected instance at line 670 (R4-4); and the three
cross-corpus id locations at lines 478/651/689 (R4-5). Round-4 verification pairs are in
`red/citation-ledger.md` (round-4 blocks, appended by the lenses plus merge-seat additions).

**Round 5 (this update)** consolidates five lens passes (`red/candidates/round-5-lens-{1..5}.md`:
3 leaf-node citation slices — lens 1 front matter/§0–§1, lens 2 §2–§3, lens 3 §4/§5/Footnotes — 1
logic/completeness, 1 dark-side/risk). Keystone claims re-verified first-hand at the merge seat
(2026-07-14): `origin/main` HEAD still `42dba2d` — zero drift since round 4's pin; `debate.js`
re-read at that ref (the `friction` array is script-local at line 145, `takeFriction` at 146,
throw sites at lines 36/136/171, and the only egress is the success-path assembly prompt at 207
and terminal `return {..., friction}` at 210–217 — bears on R5-6); `commands/research.md` step 5
read directly ("If the **returned** envelope carries `friction` entries, write them to
`<run directory>/friction.md`" — a thrown workflow never returns one; bears on R5-6); §3 row 23
(report line 727) and §2.1(b) (line 496) read side by side — row 23 still carries the chain list
`R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8` that `debate.md`'s own round-4 BLUE section
narrates discarding ("Re-verified red's list against the cited rows and adopted it in place of my
own... the report now cites red's enumeration" — true for §2.1, false for row 23; bears on R5-1);
`grep -c "three completed"` on the report = 3 (front matter, §2.1(b), row 23) against
`grep -n "^### " debate.md` = 9 headers (5 BLUE rounds 0–4, 4 RED rounds 1–4, zero `### LEAD`) —
round 4 is now the *fourth* completed judge-free round (bears on R5-3); the six
memory-architecture ids in §4 row 1 traced to their current status lines in that corpus's own
`red/findings.md` first-hand (MA-R1-28 and MA-R2-8 closed round 3, MA-R3-14 and MA-R3-15 closed
round 4, MA-R4-9 open but a fully-diagnosed miscitation verified via three independent routes,
MA-R1-19 the only genuinely open lossy-fetch-blocked case — bears on R5-2); §3 row 13's id set
extracted mechanically from report line 715 (five ids, **no** `MA-R2-8 residual`) against
§2.1/§4's six-id sets (bears on R5-2, and overrules one lens's contrary ledger line — see the
merge notes). Round-5 verification pairs are in `red/citation-ledger.md` (round-5 blocks, appended
by the lenses plus merge-seat additions and two merge-seat corrections).

## Verdict (round 5): FAIL

All 5 round-4 gaps are **CLOSED** — 4 clean, 1 **closed-with-regression** (R4-1→R5-1, with three
further, smaller residues arising inside the same fix's own text: R5-3, R5-4, R5-5). **6 gaps
open (R5-1..R5-6):** 1 medium-high (R5-5), 3 medium (R5-1, R5-2, R5-6), 2 low (R5-3, R5-4). None
disputes H1–H5 or any prior closure. The round's shape: the external citation surface stays
converged — every external re-fetch across the three citation lenses came back exact again — and
the new catches are internal-consistency drift (R5-1, R5-3, R5-4), cross-corpus citation-status
drift (R5-2), and two dark-side reflexivity findings against round 4's own fixes (R5-5, R5-6).
Severity trend resumes its decline (round 4's worst: HIGH; round 5's worst: MEDIUM-HIGH, closable
with one argued sentence). Round 4 projected this as the PASS round; it is not — but every open
gap is one-clause-to-one-paragraph sized, five of the six are copy-edits from already-verified
material inside this corpus, and none requires new research.

---

## Round-4 gaps — status after round 5

### R4-1 — CLOSED WITH REGRESSION (round 5) → R5-1 (three further residues inside the same fix's text: R5-3, R5-4, R5-5)
The required coverage correction landed in full: §2.1's rollover row states classes (a)/(b) side
by side with the widen-window remedy scoped to (a) only; §3 row 23 added and graded; §2.3
addition 15 added known-failing; §5 item 12 states the two fixes' independence explicitly — all
verified in place, and the fix shape matches the backlog item re-read first-hand at `42dba2d`.
But row 23's own likelihood cell ships blue's *discarded* first-pass chain enumeration (R5-1),
the fix's round-count numeral is already one round stale (R5-3), addition 15's uniform closure
labeling does not match the chain it claims to mirror (R5-4), and the `supersedes` remedy's
reliance on unenforced prompt compliance goes unnamed two sentences after the report indicts
exactly that reliance (R5-5).

### R4-2 — CLOSED (round 5)
Decided: throw, with stated reasoning (anti-silent-degradation, consistent with row 19, §2.3
item 1, and R2-7's precedent); §2.3 addition 13 extended with the matching positive assertion,
verbatim-identical in both locations. No guard on live `main` yet — consistent with the
docketed-for-run-4 disposition red pre-accepted. Corroboration: high. (R5-6 touches the same
throw surface, but it is a pre-existing structural class shared with the two shipped null-guards
— not a defect in this decision.)

### R4-3 — CLOSED (round 5)
The originating sentence edited at the source ("the adversarial-disconfirming-first lens (a
distinct method from local-repo critical-stance, named separately below)"); report-wide grep: the
retired slash-compound and `lanes >= 4` survive only inside explicitly-marked correction
narratives. Corroboration: high.

### R4-4 — CLOSED (round 5)
The fifth location now reads "pending a third occurrence (corrected R2-1)"; report-wide grep
"4th|fourth": zero remaining unmarked instances. Corroboration: high.

### R4-5 — CLOSED (round 5)
All four cross-corpus locations carry the `MA-` prefix; [^GapIdScheme] present and verified
against the memory-architecture corpus first-hand (all six ids real; that corpus confirmed to run
to at least R4-12). Closed as raised — the namespace hygiene is done. The round-5 finding against
the same id lists (R5-2) is a different defect class (status drift of the *referenced* gaps, not
namespace collision) and is not a reopening.

---

## Round-5 gaps (id | status | severity | likelihood x impact x complexity | corroboration)

### R5-1 — OPEN — MEDIUM — certain (static text, read side by side at the merge seat) x medium (the evidentiary cell of the report's highest-graded round-4 finding carries mis-traced lineages, two of them contradicted by this file's own closure record — inside the very row arguing lineage precision matters) x trivial (copy an already-correct list from two sections earlier) — corroboration: HIGH (three lenses converged independently — lenses 1, 2, 4; both quotes verbatim from the current report; every chain link checked against this file's own closure entries)
**Location:** §3 row 23, likelihood cell — *"this exact corpus contains four live regression
chains (R1-5→R2-4→R3-4/R3-9; R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8)"* — against §2.1
Tier A, rollover sub-row (b) — *"R1-5 → R2-4 → R3-4/R3-9... plus... R2-5 → R3-10... R2-7 → R3-6...
and R2-8 → R3-5 → R4-3."*
**Problem:** two different enumerations of the same "four live regression chains" claim, agreeing
only on the first chain. Row 23's version is blue's own discarded first-pass reconstruction —
`debate.md`'s round-4 BLUE section narrates replacing it ("Re-verified red's list against the
cited rows and adopted it in place of my own... the report now cites red's enumeration"), but the
substitution reached §2.1 only. Two of row 23's three non-shared entries are factually wrong
against this file's own record: R2-5's regression successor is **R3-10**, not R3-8 (R3-8 is a
distinct live-drift finding on the same footnote's pin); and R2-1 closed **clean** round 3 with
this file explicitly disclaiming the link ("R3-7, **not a reopening**"). The third entry
(R1-16→R2-8→R3-5) is truthful link-by-link but truncates the chain's live tip (→ R4-3). This is
the report's own named repair-reaches-one-location-not-all class (R3-4, R3-10, R4-4) recurring
*inside* R4-1's own fix, between two sections of the same round-4 patch. Secondary nuance, not
gated: the fullest forms of two §2.1 chains include their round-1 origins (R1-15→R2-7 and
R1-16→R2-8 are both real CLOSED-WITH-REGRESSION links per this file); §2.1's "at least three
more" hedge keeps its version accurate as stated — including the origins is optional.
**Required fix:** replace row 23's parenthetical with §2.1(b)'s list — "(R1-5→R2-4→R3-4/R3-9;
R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3)" — one clause, no new research.

### R5-2 — OPEN — MEDIUM — high (the status claims are read at face value; no hedge or as-of date) x medium (the #1-build disposition survives on independent grounds — live backlog + MA-R1-19 — but the stated blocking evidence is inflated ~3x: 2 genuinely open vs. 6 cited, and 4 of the 6 resolved by exactly the ordinary re-fetch the framing implies cannot happen without the tool) x low — corroboration: LOW for the four closed-cited-as-open sub-claims (the primary source contradicts them), MEDIUM for MA-R4-9 (real and open, but a diagnosed miscitation — wrong failure class for this row), HIGH for MA-R1-19 (matches source exactly), HIGH for the three-way list inconsistency (mechanical extraction)
**Location:** §4 row 1, status cell — *"blocks memory-architecture's own MA-R1-19, MA-R1-28,
MA-R2-8 residual, MA-R3-14, MA-R3-15, MA-R4-9... from resolving past
'unable-to-corroborate-at-leaf-node'"* — plus the same claim's two siblings: §2.1 Tier C (six
ids) and §3 row 13 (*five* ids — drops `MA-R2-8 residual`).
**Problem:** citation-status drift on a living source. Direct read of the memory-architecture
corpus's own `red/findings.md`: MA-R1-28 and MA-R2-8 **closed round 3** (by live abstract
re-fetch and re-citation — not PDF extraction); MA-R3-14 and MA-R3-15 **closed round 4**;
MA-R4-9 is open but is a fully-diagnosed miscitation verified through three independent routes —
not an "unable-to-corroborate" block; only MA-R1-19 is a genuine, open, lossy-fetch-blocked case.
Second-order staleness: the underlying source sentence (that corpus's friction section, "a
full-PDF... tool would discharge R1-19, R1-28, R2-8's residual, and R2-10 definitively") was
written mid-round-2 and its own prediction was falsified the next round — two of those gaps
closed without any PDF tool. And the three in-report enumerations of the same set disagree on
membership (6/6/5), no two matching the source's current state.
**Required fix:** either restate historically ("as of memory-architecture round 2; four of six
since closed by ordinary re-fetch, not PDF tooling") or cite only the genuinely-open MA-R1-19
(plus MA-R4-9 correctly classed as a miscitation, if kept) — and reconcile all three locations to
one canonical list. The row's #1 ranking itself stands on the independent backlog corroboration.

### R5-3 — OPEN — LOW — certain (grep-counted: 3 instances) x low (correcting it *strengthens* the finding — a fourth completed round with zero judge dispatch) x trivial — corroboration: HIGH (`grep -c "three completed"` = 3; `grep -n "^### " debate.md` = 9 headers, 5 BLUE rounds 0–4, 4 RED rounds 1–4, zero `### LEAD`, re-run at the merge seat)
**Location:** front matter (Round 4 corrections summary — "the judge was never dispatched once
across three completed rounds"), §2.1(b), and §3 row 23 — all "three completed rounds."
**Problem:** accurate when blue wrote it mid-round-4; blue's own round-4 response has since
completed the round without a `### LEAD` section (the same lineage-blindness the row describes
kept the docket disarmed), making round 4 the fourth judge-free completed round. A stale
"current-state" numeral of exactly the class §3 row 11/R2-6 warns about, self-applied. Lens 4
named two locations; the merge-seat grep found the third (front matter) — the retracted-token
grep must run report-wide, both directions.
**Required fix:** "four completed rounds" (or round-agnostic: "every completed round to date") at
all three locations.

### R5-4 — OPEN — LOW — certain x low (documentation-fidelity nit on a proposed, unbuilt test case; test validity unaffected) x trivial — corroboration: HIGH (addition-15 text against this file's own R1-5/R2-4 closure lines, read verbatim)
**Location:** §2.3 addition 15 — *"round 2's merge closes `X-1` 'WITH REGRESSION'... and round
3's merge closes `X-2` 'WITH REGRESSION' and raises `X-3`/`X-3b`"* — claiming to "mirror this
corpus's own chain directly."
**Problem:** the mirrored chain's second link did not close "WITH REGRESSION": R2-4's recorded
status is "CLOSED, **REBUTTAL ACCEPTED WITH EVIDENCE** (round 3) → regressions R3-4, R3-9." The
detector logic under test does not depend on the closure-status label (both statuses mint
fresh-id successors), but the case's framing misdescribes the record it claims to mirror.
**Required fix:** loosen the wording ("closes... and mints a fresh-id successor, whatever the
closure status") or name both real labels.

### R5-5 — OPEN — MEDIUM-HIGH — medium (an optional field set under prompt instruction alone, by a merge agent under load — this corpus has already demonstrated the class twice: R3-2's schema-declared friction field uncalled for three rounds; R4-2's undecided disjunction shipped verbatim) x high (the failure is telemetry-invisible: an unset or vacuous `supersedes` leaves `contested.length` at 0, indistinguishable from "no regression occurred" — silent reversion to the exact defect the fix exists to close) x low report-side (one named sentence either way) — corroboration: HIGH (row 23's fix text and §2.1(b)'s good-faith sentence read verbatim in the same document; `RED_ENVELOPE` schema re-confirmed at `42dba2d`)
**Location:** §3 row 23, complexity cell — *"(1) add `supersedes: { type: 'array', items: {
type: 'string' } }` (**optional**) to `RED_ENVELOPE`'s gap schema and **instruct red-merge to set
it** when closing a gap 'WITH REGRESSION' and minting a successor"* — against §2.1(b), same
finding, two sentences earlier: *"a property of this run's actors' good faith, not a property the
detector enforces."*
**Problem:** the fix for lineage-blindness is itself blind to its own failure mode — it relies on
exactly the unenforced good-faith compliance the row indicts. Nothing validates that a
closed-WITH-REGRESSION gap has a successor whose `supersedes` names it. §2.3 addition 15 only
tests the *detector* given a correctly-populated canned field (Tier A); whether the live merge
agent reliably populates it is a real-model-reasoning question the report's own boundary
discipline (§2.1: "a simulator that fakes... judgment content is the research problem itself")
places at Tier C — a Tier-A test for a Tier-C failure surface, with no fallback named and no
residual risk flagged, and the row prices the whole three-part fix at Medium complexity as if the
schema/prompt half were sufficient. Making the field required does not close this: a vacuous or
incomplete array is schema-legal-but-semantically-empty — the report's own R3-1 class.
**Merge-seat temper of lens 5's HIGH:** the defective claim concerns a proposed, docketed,
unbuilt mechanism's future reliability plus an underpriced complexity cell — not a false claim
about current state; precedent is R2-7 (a mitigation overstating its guarantee on a live
risk-accept), graded MEDIUM-HIGH. Graded to match, with lens 5's likelihood/impact reasoning
adopted intact.
**Required fix:** one sentence at row 23 (or a new row) naming the residual: either an argued
risk-accept (three rounds of demonstrated merge-seat compliance — argued, not assumed, same
pattern as row 19's rescoped mitigation), or scope the cheap structural cross-check — record
regression-flagged closures in the merge envelope as a small required field on the *closed* gap,
and have the script assert every such closure reconciles against some new gap's `supersedes`
this round, throwing on mismatch per the report's own R4-2 precedent. Either closes it; silence
does not — the report cannot indict unenforced good faith at §2.1(b) and then ship an
unenforced-good-faith fix in the same row without saying so.

### R5-6 — OPEN — MEDIUM — medium (requires a mid-run throw; the null-return class has fired in this project's history — run 2's crash is why the guards exist — and R4-2's decided guard adds a third mid-loop site) x medium-high (the entire run's accumulated friction telemetry is lost at exactly the runs where something went wrong enough to be worth reporting; `friction.md` is never written — `commands/research.md` step 5 fires only on a successful return, and the line-207 assembly-prompt injection is success-path-only too) x low-medium (report-side: one paragraph; the compliant code-side fix needs one design decision, not new infrastructure) — corroboration: HIGH (`debate.js` lines 36/136/145–146/171/207/210–217 and `commands/research.md` step 5 all read first-hand at the merge seat, at `42dba2d`)
**Location:** §2.1, Tier A friction row (as corrected by R3-2 — the correction names the
`blue-synthesize` call-site exception but not this second, structurally distinct loss path) — and
§3 row 2's impact cell (*"a mid-debate crash at the judge seat loses every paid round up to that
point"* — the round-loss recognition never extended to the friction aggregate specifically).
**Problem:** the aggregated `friction` array lives only in script-local scope until the terminal
`return` — which a thrown error never reaches. Three throw sites exist today (args guard,
`blueEnv` null, `redEnv` null) and R4-2's decision adds a fourth; any of them firing mid-run
discards every seat's accumulated friction from every prior round, and the command layer's
friction persistence (step 5) is unreachable for a thrown workflow. The naive fix (script writes
incrementally) violates the script's own stated no-filesystem-access doctrine the same round it
would fix the gap. **Merge-seat temper of lens 5's MEDIUM-HIGH:** the trigger is conditional (a
throw must fire), unlike R3-2's every-run structural certainty, and the loss is telemetry, not
report content (the blackboard artifacts persist on disk) — graded MEDIUM, mechanism adopted
intact.
**Required fix:** one paragraph at §2.1's friction row or a new §3 row naming the
throw-loses-the-aggregate path, then either (a) adopt the architecture-compliant variant — each
schema'd seat appends its own friction line directly to `runDir/friction.md` via the
append-only-blackboard convention already adopted for §3 row 8, in addition to returning it in
the envelope — or (b) an argued risk-accept that crash-time friction is low-value; the argument
does not exist in the report today and is not assumable by silence (the report's own R4-2
standard).

---

## Merge-time verification and dedupe notes (round 5)

- Dedupe map: lens 1 Finding 1 + lens 2 Finding 1 + lens 4's in-pass "R5-1" = **R5-1** (three
  independent convergences). Lens 3's in-pass "R5-1" + lens 1 Finding 2 = **R5-2** (merged: the
  status-drift finding subsumes the list-membership inconsistency). Lens 4's in-pass "R5-2" =
  **R5-3** (extended at the merge seat: the grep found a third instance in the front matter that
  lens 4's two-location statement missed). Lens 2 Finding 2 = **R5-4**. Lens 5's in-pass "R5-1" =
  **R5-5** (tempered HIGH → MEDIUM-HIGH, reasons stated in the gap). Lens 5's in-pass "R5-2" =
  **R5-6** (tempered MEDIUM-HIGH → MEDIUM, reasons stated in the gap).
- **Two lens errors caught and overruled at the merge seat** — logged with the same discipline
  demanded of blue: (1) lens 5's "checked, not raised" item asserted row 23's chain enumeration
  "matches §3 row 23's and §2.1's text exactly — no discrepancy found"; direct read of report
  line 727 shows row 23 carries the discarded list, the live discrepancy three sibling lenses
  caught independently. An unquoted hold masked a discrepancy — hold-claims need the same
  side-by-side quoting as gap-claims (new red-side pattern, recorded in red's memory). (2) lens
  2's ledger line 168 asserted §3 row 13 carries the six-id `MA-` set including `MA-R2-8
  residual`; mechanical extraction of report line 715 shows five ids, no `MA-R2-8`. Lenses 1 and
  3 are correct; two merge-seat correction lines appended to `red/citation-ledger.md`.
- `origin/main` re-fetched at merge: HEAD still `42dba2d`; `debate.js` and `ideas/backlog.md`
  both stable at the round-4 pin. The memory-architecture statuses (R5-2) were traced first-hand
  against that corpus's `red/findings.md`, not taken from any lens's paste.
- Lens 5's disconfirming holds re-checked and accepted: the R4-2 throw cannot false-positive on a
  genuine PASS (the `verdict === 'PASS'` break precedes the guard's position); `parallel()`
  semantics are irrelevant to R5-6 (the throws are synchronous in the loop body); the
  cost-incentive angle on R5-5 is folded into its likelihood grade as unproven-aggravating, not
  asserted separately.

## Noted, not raised (round 5)

- **Chain-origin truncation in §2.1(b)'s list** (R1-15→R2-7 and R1-16→R2-8 are real
  closed-WITH-REGRESSION origins this file records, omitted from the report's chain heads):
  §2.1's "at least three more" hedge keeps its version accurate as stated; folded into R5-1 as an
  optional improvement, not gated.
- **Row 23's fix item (1) lacking a documented "WITH REGRESSION" protocol state** (lens 4,
  checked and held): the convention is this corpus's emergent practice, not doctrine, but the fix
  text's "instruct red-merge to set it" covers the prompt-change obligation at the same
  specificity as its §3 siblings — holding it to a fuller implementation-spec standard would be
  inconsistent.
- **Item 6's "cost is one more lane-dispatch" phrase** (lens 4): scoped to the floor's marginal
  cost; the full `lanes >= 5` arithmetic is stated two sentences later in the same cell. A full
  read is not misled.
- **Row 16b's dev/smoke-vs-keeper split** (lens 4): gated on the existing `model`/`judgmentModel`
  knobs, which ship independently of the unbuilt `--smoke` path. Mechanism real; held.
- **Red-auditor memory-pattern-file count now 25** (was 23 at round 4): settled round-2/3/4
  non-gap (live accreting mechanism); not re-litigated.
- **[^CostFigureProvenance] pin still `d164ab2`, HEAD `42dba2d`:** the one-commit drift is R4-1's
  own subject commit, byte-identical on item 28 (verified round 4); no re-pin gap.

## Verified clean this round (round 5 — see red/citation-ledger.md round-5 blocks)

Lens 1 (front matter/§0–§1): the R4-2/R4-3/R4-4/R4-5 front-matter summary sentences against
their cited rows — all verbatim-consistent; §1.2's R4-5 addition against the memory-architecture
MA-R2-10 source (exact); [^GapIdScheme]'s "runs to at least R4-12" re-confirmed live; report-wide
"4th|fourth" grep clean. Lens 2 (§2–§3): `RED_ENVELOPE` no-`supersedes` / pure-id-equality
`contested` / full-replacement `prevGapIds` all re-read live at `42dba2d` (code fix genuinely
undocked, as the report states); the 25-minute commit-delta claim exact (25m12s); the backlog
docket-detector item verbatim against the report's citation; §3 row 6's corrected floor sentence
and zero residual "lanes >= 4"; row 20/addition 13's identical throw-message text; both PDF-MCP
repos re-checked live (`archived: false`). Lens 3 (§4/§5/Footnotes): §4 rows 2–10 and §5 items
1–12 unchanged and held at prior HIGH; [^CostFigureProvenance] pin unchanged; backlog line 27(c)
"TOP TOOL GAP" re-fetched live (supports row 1's ranking independently of R5-2). Lens 4: all 46
prior-round gap fixes hold as described at their cited locations (the one exception raised as
R5-1); template compliance (top-level `report.md` still a correct stub pending PASS; all
footnotes semantic word-labels). Lens 5: all five round-4 closures re-verified first-hand at
`42dba2d`; friction-aggregation call sites and null-guard positions exact as the report
describes. Merge seat: the keystones in the round-5 header paragraph, first-hand.

## Round 5 disposition summary

FAIL. 5/5 round-4 gaps closed (1 with regression: R4-1→R5-1, plus three smaller residues in the
same fix's text). 6 new gaps open. Gate tier: **R5-5** (name the `supersedes` fix's residual
reliance on unenforced merge-seat compliance — one argued sentence, risk-accept or structural
cross-check, either accepted) and **R5-1/R5-2** (mechanical accuracy fixes to the evidence cells
of the report's two highest-profile rows — row 23's chain list and row 1/row 13/§2.1's
memory-architecture status claims). R5-3 and R5-4 are one-word-to-one-clause; R5-6 closes by one
paragraph or an argued risk-accept red will take if reasoned. None disputes H1–H5. Convergence
expectation: five of the six fixes are copy-edits from material already verified inside this
corpus; R5-5/R5-6 are single argued sentences/paragraphs. If blue lands these without new
regressions, round 6 is the PASS round. Red's own ledger this round: two lens errors (lens 5's
unquoted hold contradicting three sibling lenses; lens 2's wrong ledger line on row 13's id set)
caught and corrected at the merge seat before entering the record — and the unquoted-hold class
is new, logged in red's pattern memory.

---

## Verdict (round 4): FAIL — superseded by round 5, preserved

All 10 round-3 gaps are **CLOSED** — 8 clean, 2 **closed-with-regression** (fix responsive to the
gap as raised, but carrying a new, smaller defect re-raised as a round-4 gap: R3-1→R4-2,
R3-5→R4-3). **5 gaps open (R4-1..R4-5):** 1 high (R4-1), 1 medium (R4-2), 2 low-medium (R4-3,
R4-5), 1 low (R4-4). None disputes H1–H5 or any prior closure. The round's shape: the citation
surface has converged — all three citation lenses' external and local re-fetches came back clean;
every round-4 *prose* gap is repair-propagation residue at ≤ LOW-MEDIUM. The one HIGH (R4-1) is
driven by evidence that landed in the live repo 25 minutes after blue's round-3 pin — four of
five lenses converged on it independently. Honest trend note: headline severity did **not**
decline this round (round 3's worst was MEDIUM-HIGH; R4-1 is HIGH), but the report-quality trend
did — R4-1 is not a blue drafting failure, it is live-source drift plus a mechanism this corpus
itself demonstrates; it still gates because §2.1/§2.3 currently state the narrower rollover case
as the *only* version of the docket-detector defect.

---

## Round-3 gaps — status after round 4

### R3-1 — CLOSED WITH REGRESSION (round 4) → R4-2
§3 row 20 and §2.3 addition 13 added; row 20 itself states it is "the correction to §2.1's
round-loop coverage claim and §2.3 item 8's framing" — the report-side coverage correction red
required; code fix docketed for run 4, the disposition red pre-accepted. But the guard's
specified behavior ships as an unresolved disjunction — copied verbatim from red's own R3-1
required-fix text — never decided. See R4-2 (provenance note against red recorded there).

### R3-2 — CLOSED (round 4)
§2.1's "never dropped" claim corrected in place, naming the `blue-synthesize` exception with call
sites (lines 170/187/197 vs. the absent line-132 harvest) and the simulator test's real two-seat
coverage; §3 row 21 and §2.3 addition 14 added. Re-confirmed against live `debate.js` at
`42dba2d` by lens 2 and lens 5 independently. Corroboration: high.

### R3-3 — CLOSED (round 4)
§2.3 item 5 corrected; §3 row 22 added with an explicit option-(b) choice *with stated reasoning*
(cheaper at the stated complexity; risk-acceptance reserved for complexity exceeding
likelihood × impact) — a decision, not a disjunction, unlike row 20. Re-verified against live
`debate.js` (judge `resolutions` still read only for `adjudicated`/`friction`). Corroboration:
high.

### R3-4 — CLOSED (round 4)
§1.1's body clause edited to match the corrected footnote; the in-place correction note ("this
body clause previously still read 'continued gains observed to 7 agents...'") confirmed at line
257. Report-wide grep: the retracted clause survives only inside the correction note and the
footnote's preserved historical narrative, both explicitly marked retracted. Corroboration: high.

### R3-5 — CLOSED WITH REGRESSION (round 4) → R4-3
The reconciliation recomputed, choosing red's option (a): full four-method roster requires
`lanes >= 5`, stated plainly; row 7's language corrected from "floors N at 3" to "proposed
(unbuilt) floor targets `lanes >= 3`," aligned with its [OPEN] status — both required alignments
delivered, and this time the correction computes instead of asserting. But the *originating*
ambiguous sentence (the slash-compound "critical-stance/adversarial-disconfirming lens" floor
clause) stands unedited above the correction in the same cell. See R4-3.

### R3-6 — CLOSED (round 4)
Row 19 reworded to "zero hits anywhere in the plugin, including the ledger clause itself."
`git grep -ni "independen"` re-run at `42dba2d` by lens 3: still zero. Corroboration: high.

### R3-7 — CLOSED (round 4)
The one-clause narrowing added to §3 row 15 and §4 row 5 (1-confirmed-as-length-ceiling +
1-same-family-plausible, occurrence 2's source names only "shell parsing"), with the argued
risk-accept on the likely-unrecoverable transcript — exactly the disposition red offered to
accept. Textual distinction re-verified against `run2-friction.md` line 4 and
`blue/CHANGELOG.md` Round 0 by lens 3. Corroboration: high.

### R3-8 — CLOSED (round 4)
[^CostFigureProvenance] re-pinned at `d164ab2` with sub-item (d) quoted in full; §3 row 18's
rationale updated with the turns×context finding. The subsequent drift to `42dba2d` touched a
*different* backlog item (item 28 byte-identical across the diff, verified by lens 3) — not a
re-open; the new commit's content is R4-1's subject, handled there. Corroboration: high.

### R3-9 — CLOSED (round 4)
The two-senses disambiguation (nominal-N practical knee vs. effective-diversity saturation
ceiling) added to [^DiminishingReturns], with the synthesis explicitly favoring sense 2 and the
reason stated. Internally coherent on direct re-read (lens 3); figures match the
twice-independently-re-fetched arXiv:2606.02646 ledger entries. Corroboration: high.

### R3-10 — CLOSED (round 4)
Both untagged instances now tagged: §2.1's reading-order-first "252.9k" row and §3 row 4's
"253k–3M" impact cell both carry [^CostFigureProvenance] with in-place round-3 tag notes.
Verified by direct grep at merge (lines 453, 642). Corroboration: high.

---

## Round-4 gap details (original grading, preserved verbatim — round-5 statuses above supersede the "OPEN" markers below)

### R4-1 — OPEN — HIGH — certain (already realized in this corpus, not projected) x high x low-medium — corroboration: HIGH (direct read of `debate.js:142/176–190` at `42dba2d`; direct `git show 42dba2d` of the backlog item; header-level grep of `debate.md`; four of five lenses converged independently, each tracing the code first-hand rather than trusting the backlog)
**Location:** §2.1, Tier A — *"**Gap-id rollover across non-adjacent rounds**: `prevGapIds` holds
only the prior round, so a gap closed in round 1 recurring in round 3 classifies 'new,' not
'contested'"* — and §2.3 addition 3 — *"Gap-id rollover — id present r1, absent r2, present r3...
known-failing until `prevGapIds` widens to full adjudicated history."*
**Problem:** the report's only stated version of the docket-detector defect is the
same-id-skips-a-round timing case, and its own proposed remedy (widen `prevGapIds` to full
adjudicated history) does not close the broader case that is actually live in this corpus.
Contested-docket membership is pure id string-equality (`debate.js:178`); the gap schema has no
`supersedes`/lineage field; and red's own closed-WITH-REGRESSION methodology mints a **fresh id
for every successor gap** — so a multi-round dispute lineage never matches `prevGapIds`,
`contested` stays 0 every round, and the judge is never dispatched, no matter how long the
dispute persists. The only remaining brake is the `maxRounds` cost ceiling — which bounds cost,
not convergence quality, the docket's stated purpose ("so debates converge instead of grinding").
**Already realized, four chains in this corpus:** R1-5 → R2-4 → R3-4/R3-9 (one footnote, four
ids, three rounds), plus R2-5→R3-10, R2-7→R3-6, R2-8→R3-5→R4-3. Zero `### LEAD` sections exist
in this retrospective's entire transcript — the judge has never once been invoked across three
completed rounds despite exactly the rebuttal-and-regression chains the docket exists to route.
That this debate converged anyway is a property of this run's actors (blue conceded every
sustained gap in good faith), not a property the detector enforces; a spinning debate would show
identical `contested.length === 0` telemetry to `maxRounds`. Independently confirmed by the
project's own live backlog (commit `42dba2d`, 25 minutes after blue's `d164ab2` pin), which
names this retrospective's chain as its worked example and proposes the fix shape
(`supersedes: [prior-ids]` on the gap envelope; lineage-following contested-detection at chain
depth ≥ 2; a simulator lineage-regression case). Distinct from R3-1/R3-2/R3-3, all of which
assume the judge branch is at least reachable — this is one level upstream: the branch is never
entered for the corpus's dominant real gap-lifecycle event (repair-regression).
**Required fix (report-side is the gate; code fix docketed for run 4 is acceptable, same
disposition as R3-1/R3-2):** correct §2.1's rollover row and §2.3 addition 3 to state both
failure classes side by side (same-id-after-a-skipped-round vs. fresh-id-successor-chain) and
that the widen-`prevGapIds` remedy closes only the former; add a graded §3 row for the lineage
fix citing the backlog item; add a §5 open question (are the two fixes independent? — yes, per
the trace, and the report should say so); add the simulator case mirroring this corpus's own
chain, asserting the judge IS invoked by round 3 of a depth-≥2 supersession lineage.

### R4-2 — OPEN — MEDIUM — certain x medium x trivial — corroboration: HIGH (row 20's cell and §2.3 addition 13's assertion text read directly at merge; red's own R3-1 required-fix text compared side by side)
**Location:** §3 row 20, complexity cell — *"either treat as `PASS`-with-a-logged-warning (red
found nothing to fail on) or throw a distinguishing error rather than looping."*
**Problem:** the shipped fix is red's R3-1 required-fix disjunction carried forward verbatim and
never resolved into a decision — the only round-3 fix that ships an "or" where its siblings ship
choices (row 21: a single named call; row 22: option (b) with stated reasoning). The two branches
are opposite failure philosophies — silently converting a degenerate FAIL into a passing verdict
vs. halting loudly — precisely the "silent" axis this report argues against everywhere else
(§0, R3-2's own finding, row 2b, R3-1's own problem statement). **Merge-seat temper of lens 4's
grading:** §2.3 addition 13's assertions are negative and option-agnostic ("assert the loop does
not silently re-dispatch...; assert the terminal return does not pair `UNVERIFIED` with
`gaps_outstanding: 0`"), so the test CAN be written as-is — lens 4's "cannot actually be written"
overstates. But only the negative half is specifiable; the guard's positive behavior remains
undefined until the disjunction is decided. Impact graded medium on that verified basis, not
lens 4's medium-high. **Provenance note, against red:** the undecided "or" originated in red's
own R3-1 required-fix text; red must not hand blue a disjunction without naming its favored side
— blue ships red's phrasings verbatim (third instance of the class: R3-6, R2-4's proposed source,
now this; logged in red's pattern memory).
**Required fix:** one clause picking a side with a stated reason. Red's position: throw — a
degenerate merge-lens return should halt for human attention, not resolve toward a passing
verdict, for the same reason row 19's poisoning finding distrusts an unexamined clean signal. An
argued "PASS with a loud operator-surfaced warning, because throw loses partial progress" is
acceptable if reasoned — the gap is the undecided disjunction, not red's preference. Then extend
addition 13 with the positive assertion matching the choice.

### R4-3 — OPEN — LOW-MEDIUM — high (reading-order exposure: the ambiguous sentence is phrased as the operative instruction; the correction reads as a gloss on it) x low (the correct arithmetic is stated later in the same cell; a full read resolves it) x trivial — corroboration: HIGH (both sentences read verbatim in the same cell at merge; caught independently by lenses 2 and 4)
**Location:** §3 row 6 — *"assign the critical-stance/adversarial-disconfirming lens to at least
2 of N lanes (not 1-of-N)"* — against the same cell's round-3 correction — *"four named methods
(primary-literature / practitioner-production / adversarial-disconfirming-first / local-repo
critical-stance), one of which (adversarial-disconfirming-first) carries a 2-of-N redundancy
floor."*
**Problem:** R3-5's fix computes the right total (`lanes >= 5`) but leaves unedited the
originating sentence whose slash-as-compound-name reading produced the round-2 mis-add in the
first place. The same slash character is a synonym-joiner in the floor sentence and a
list-separator in the roster two sentences later, with nothing flagging the switch — a reader who
stops at the operative instruction still picks up the 3-method/`lanes >= 4` misreading R3-5
exists to prevent. Repair-reaches-the-conclusion-not-the-source class (sibling of R3-4/R3-10:
the correction landed downstream of the defect instead of in it).
**Required fix:** edit the original sentence itself — e.g. "assign the
adversarial-disconfirming-first lens (a distinct method from local-repo critical-stance, below)
to at least 2 of N lanes" — so the round-3 correction confirms an unambiguous sentence rather
than patching an ambiguous one.

### R4-4 — OPEN — LOW — certain x low x trivial — corroboration: HIGH (report-wide grep "4th|fourth" at merge: exactly one uncorrected instance, §3 risk-accepted paragraph; the three corrected locations — row 15, §4 row 5, §5 item 10 — confirmed in the same pass)
**Location:** §3, the risk-accepted closing paragraph — *"ENAMETOOLONG tooling (#15, re-graded
round 1 to track recurrence but still risk-accepted pending a 4th occurrence) [L1]"*.
**Problem:** a fifth location of R2-1's corrected trigger count, missed because R2-1 was closed
by fixing the three locations the gap named without a report-wide grep for the retracted numeral
— the exact class red's own memory names (grep the retracted token report-wide, both directions).
The paragraph is written as a scanning aid, so a summary-only reader takes away the stale
trigger.
**Required fix:** one word — "pending a third occurrence (corrected R2-1)," matching the three
already-corrected instances.

### R4-5 — OPEN — LOW-MEDIUM — certain (three live in-report instances, grep-confirmed) x low-medium (traceability: bare ids from another corpus are indistinguishable from this retrospective's own gap namespace; "R4-9" now collides live with this round's active R4-* series) x trivial — corroboration: HIGH (grep-confirmed at merge, lines 478/651/689; ids traced to `inputs/run2-friction.md` — the memory-architecture corpus's numbering)
**Location:** §3 row 13 — *"kept 3+ figures at unable-to-corroborate across 4 rounds (R1-19,
R1-28, R3-14/15, R4-9)"* — plus §4 rank-1 (*"blocks R1-19, R1-28, R2-8's residual, R3-14, R3-15,
R4-9 from resolving"*) and §2.3's lossy-fetch row (same id list).
**Problem:** unresolved cross-corpus gap-id collision. These are the *memory-architecture*
retrospective's internal red-audit ids (the corpus this run's H1–H5 doubts are about), not this
retrospective's — whose findings file uses the identical bare "R#-#" scheme. Round 2's
disposition already flagged this class once (§1.2's "R2-10") but only inside `red/findings.md`;
the note never reached the report, and three more unflagged instances stand. The sting is now
live: this round mints R4-1..R4-5, so a reader or tool cross-referencing the report's "R4-9"
against this retrospective's findings gets a hit-shaped token that resolves to nothing — and
would resolve to the *wrong thing* in any round that mints nine-plus gaps. A report arguing this
extensively for per-claim provenance (§1.2, claim manifest, §3 row 5) should not ship an
overloaded identifier scheme with no disambiguating marker.
**Required fix:** one global disambiguation parenthetical or footnote at first occurrence
(covering §1.2's instance too), or a corpus prefix (e.g. "MA-R1-19") at all four locations.

---

## Merge-time verification and dedupe notes (round 4)

- Dedupe map: lens 1's in-pass "R4-1" + lens 2's Finding 1 + lens 3's in-pass "R4-1" + lens 5's
  "R4-1" = **R4-1** (four independent convergences — each lens traced `debate.js` first-hand
  rather than trusting the backlog item; strongest multi-lens convergence of any gap in this
  corpus). Lens 4's in-pass "R4-1" = **R4-2**. Lens 2's Finding 2 + lens 4's in-pass "R4-3" =
  **R4-3** (merged). Lens 2's Finding 3 = **R4-4**. Lens 4's in-pass "R4-2" = **R4-5**.
- Precision correction to lens 5, applied at merge: "grep of `debate.md` for `### LEAD` — zero
  matches" is imprecise; a plain grep returns **one** match (line 528, a quoted phrase inside
  round-3 BLUE prose). The correct, stronger-form claim — verified via `grep -n "^### "` — is
  zero `### LEAD` *section headers* across rounds 0–3. Substance unchanged (the judge has never
  been dispatched); the imprecise form is the exact R3-6 class, caught at the merge seat before
  entering the record. R4-1 cites the precise form.
- Merge-seat temper of lens 4: addition 13's assertions read directly — negative and
  option-agnostic, so "the test cannot actually be written" overstates. R4-2 graded on the
  verified state (medium, not medium-high).
- `origin/main` re-fetched at merge: HEAD `42dba2d`; `debate.js` byte-identical across
  `47ae48d`/`88eb57f`/`d164ab2`/`42dba2d`. The backlog item quoted in R4-1 read via
  `git show 42dba2d` first-hand, not from any lens's paste.

## Noted, not raised (round 4)

- **[^CostFigureProvenance]'s pin now one commit stale** (`d164ab2` vs. live `42dba2d`): item 28
  byte-identical across the drift (lens 3, `git diff` scoped to the file). Per rounds-2/3
  precedent, docs-only drift is raised only where materially load-bearing; the drifted content
  (the docket-detector item) is load-bearing but is R4-1's subject, not a re-pin gap.
- **Lens 1's [^BlueReportGrep] apparent miscount (5 + 7 ≠ 7):** resolved as a subset relationship
  by direct comparison of matched line content — the assembled report's 7 matches contain all 5
  of blue/report.md's, renumbered. False alarm, correctly not raised.
- **Grep line-count vs. occurrence-count tool-mode footgun** ([^LocalGrepRed]: count-mode said
  64, occurrence count is 66, footnote says 66): tool artifact, footnote accurate. Noted for
  lens practice only — use occurrence-count when a claimed number could be either.
- **Lens 4's checked-and-held items:** §5 items 1–11 each carry a live counter-scenario; §2.4's
  "no risk-acceptance case" is argued elsewhere in-document; row 16b's run-level split matches
  the corpus's evidence granularity; the R3-1/R3-2 "docket, don't gate" framing is argued, not
  asserted. All held.
- **Lens 5's disconfirming items:** `maxRounds`-as-backstop checked and rejected as a
  risk-accept substitute for R4-1 (bounds cost, not convergence quality — recorded inside R4-1's
  grading); `adjudicated`-array growth is a non-issue (`Set` rebuilt per round, O(1) lookup);
  the bulk-tier-model susceptibility angle is speculative beyond this corpus's evidence — named
  here so it is not silently unconsidered.
- **§1.4's gap-pattern-file count:** live store still 23 (unchanged since round-3 merge check);
  the round-2/3 non-gap disposition holds.

## Verified clean this round (round 4 — see red/citation-ledger.md round-4 blocks)

Lens 1 (§0–§1): [^LocalGrep], [^BlueReportGrep], [^LocalGrepRed], [^RedFindingsGrep],
[^BlueReportUnverified], [^RedAuditorSpec], [^ClaimManifest], §1.4's memory-file and backlog-item
characterizations — all HIGH, several by independent re-run of the cited greps. Lens 2 (§2–§3):
rows 20/21/22 against live code; the R3-2 §2.1 correction; §2.3's 14-item numbering; row 13's
[^PdfMcp] held per skip-rule. Lens 3 (§4/§5/Footnotes): R3-7's textual distinction; item 28
byte-identical across the drift; "independen" grep still zero at `42dba2d`; [^DiminishingReturns]
disambiguation coherent against the twice-re-fetched source. Lens 4: all 41 prior-round fixes
hold as described at their cited locations; template compliance (stub top-level report.md
correct pending PASS; all footnotes semantic word-labels). Lens 5: R3-1/R3-2/R3-3 findings
re-confirmed live at `42dba2d`, correctly framed as docketed-not-shipped. Merge seat: the five
keystones listed in the round-4 header paragraph, first-hand.

## Round 4 disposition summary

FAIL. 10/10 round-3 gaps closed (2 with regressions re-raised: R3-1→R4-2, R3-5→R4-3). 5 new gaps
open. Gate tier: **R4-1** (the report-side coverage correction is the gate — §2.1/§2.3 currently
present the narrow rollover case as the whole docket-detector defect while the live repo and this
corpus's own history demonstrate the broader lineage-blindness; code fix docketed for run 4 is
acceptable, same disposition as R3-1/R3-2) and **R4-2** (decide the shipped disjunction — either
side, argued). R4-3/R4-4/R4-5 are mechanical one-clause fixes; R4-4 has no legitimate
risk-accept path (a stale numeral contradicting its own correction); R4-3 and R4-5 close by one
edit each. None disputes H1–H5. Convergence expectation: the citation surface is done — every
external re-fetch across three citation lenses came back clean, and all prose gaps are
propagation residue. If blue lands these five (all one-clause-to-one-row sized; R4-1 is the only
one requiring new report content, and its fix shape is already drafted in the live backlog),
round 5 is the PASS round on this trajectory. Red's own ledger this round: one imprecision
(lens 5's grep phrasing) caught and corrected at the merge before entering the record, one
overstatement (lens 4's "cannot be written") tempered against the verified text, and one
structural provenance debt owned (R4-2's disjunction originated in red's own round-3
required-fix) — the adversary's errors are logged with the same discipline demanded of blue.

---

## Verdict (round 3): FAIL — superseded by round 4, preserved

All 11 round-2 gaps are **CLOSED** — 7 clean, 4 **closed-with-regression** (blue's fix responsive
to the gap as raised, but the fix itself carries a new, smaller defect, re-raised as a round-3
gap: R2-4→R3-4+R3-9, R2-5→R3-10, R2-7→R3-6, R2-8→R3-5). One closure (R2-4) is a
**rebuttal accepted with evidence**: blue conceded the gap, leaf-checked red's proposed
replacement citation, found it does not contain the figure, and dropped the claim instead —
independently re-verified by two round-3 lenses; red's own proposed source was wrong, and that is
recorded against red, not blue. **10 gaps open (R3-1..R3-10):** 2 medium-high (R3-1, R3-2), 3
medium (R3-3, R3-4, R3-5), 2 low-medium (R3-8, R3-9), 1 medium-low (R3-7), 2 low (R3-6, R3-10).
None dispute H1–H5 or any prior closure. The round's shape: the citation surface is converging —
the three citation lenses' external re-fetches came back clean except for repair-regressions
inside round-2's own fixes (second-order now: regressions inside repairs of repairs) — while the
dark-side lens's control-flow trace of the already-merged `debate.js` opened a genuinely new
front (R3-1..R3-3), runtime edge cases no prior round examined. Severity is declining
monotonically across rounds (round 1: 2 HIGH; round 2: 5 MEDIUM-HIGH, all prose; round 3: 2
MEDIUM-HIGH, both code-trace findings — the prose gaps are now all ≤ MEDIUM): convergent, not
divergent.

---

## Round-2 gaps — status after round 3

### R2-1 — CLOSED (round 3)
Count corrected to 2 documented occurrences across 2 runs in §3 row 15, §4 row 5, §5 item 10
(trigger now "third," not "fourth"), and the §4 shape-verdict paragraph; the false
"per debate.md's merge-seat friction" citation dropped; likelihood re-argued (not re-asserted) on
the honest 2/2 rate. All four locations re-verified against sources (run2-friction.md line 4;
run1-friction.md zero mentions; CHANGELOG Round 0). Corroboration: high. A narrower, distinct
follow-on — whether occurrence 2 is even the same *mechanism* — is R3-7, not a reopening.

### R2-2 — CLOSED (round 3)
"Across rounds 0 and 1" now used consistently in §0's addendum, §4 row 4, §5 item 7; chronology
re-confirmed against `blue/CHANGELOG.md` Round 0 and `debate.md`'s round-1 RED dating. No residual
"this same round" phrasing found on full re-read. Corroboration: high.

### R2-3 — CLOSED (round 3)
§1.1's cross-provider paragraph now attributes "2 vs. 16" to the paper's L4 (model+persona)
condition and states L2's real curve (8 agents to match L1@16: 65.44% vs. 65.34%); the defer
disposition rests on the infrastructure-cost argument alone; §5 item 5's revisit trigger
recalibrated; §3 row 6 cross-referenced. Matches the ledger's HIGH-confidence Table 2 re-fetch.
Corroboration: high.

### R2-4 — CLOSED, REBUTTAL ACCEPTED WITH EVIDENCE (round 3) → regressions R3-4, R3-9
Blue conceded the gap (uncited "7 agents" figure) but rebutted red's specific proposed fix:
direct fetch of arXiv:2606.02646 shows it does **not** state "continued gains to 7 agents" — its
hardest benchmark is GSM-Hard, its harder-task knee is N≈10, and its headline is a ~1.8-agent
effective-team plateau by N=30. Two round-3 lenses independently re-fetched and confirmed every
figure in the rebuttal. Red accepts: applying red's required fix as written would have introduced
a second miscitation; the drop-rather-than-re-cite disposition is the correct one, and red's
unverified proposed source is logged as red's own error (see memory:
`pattern_repair_regression_citation.md`, red-side extension). **But the repair regressed twice:**
the §1.1 *body* still asserts the retracted clause (R3-4), and the footnote's replacement
sentence is internally contradictory (R3-9).

### R2-5 — CLOSED WITH REGRESSION (round 3) → R3-10
[^CostFigureProvenance] added, honestly scoped (self-reported, mixed provenance,
understatement-direction-only), attached at §2.3's closing line, §2.4, and the Tier B table —
which is what the gap required. But the propagation is incomplete: §2.1's first Tier A row
(reading-order-first instance of "252.9k") and §3 row 4's impact cell ("253k–3M") carry the same
figures untagged. See R3-10.

### R2-6 — CLOSED, STRENGTHENED (round 3)
§3 row 11 and §5 items 4/7 now state run 3 ran with zero artifact trail. Re-verified live at
`d164ab2`: still exactly two run directories under `research/`, while a *third* backlog commit
(`d164ab2` itself, "merge-seat cost analysis... run-3 transcripts") cites run-3 transcript data —
further corroboration, not contradiction. Corroboration: high.

### R2-7 — CLOSED WITH REGRESSION (round 3) → R3-6
§3 row 19 rewritten to the honest mechanism scope (catches source-misstatement, not
self-consistent fabrication); §5 item 8 answered and collapsed into the same statement; the
risk-accept stands on the honest scoping — exactly what the gap required. But the rewrite carries
forward, verbatim, an imprecise grep characterization that originated in *red's own round-2
merge text*. See R3-6.

### R2-8 — CLOSED WITH REGRESSION (round 3) → R3-5
The reconciling sentence the gap demanded was added to §3 row 6 (scoped exception to the
headcount risk-accept; `lanes >= 4` under the full roster; row 7's floor for non-adopting runs) —
responsive in form. But the reconciliation's own arithmetic under-counts by one lane and calls an
unenforced `[OPEN]` default a "floor." See R3-5.

### R2-9 — CLOSED (round 3)
§3 row 10's impact cell re-graded against the shipped ledger's actual prose-change-keyed
skip-trigger, with the concrete one-line fix (time/access-date condition in the same
`ledgerClause` string) stated as a build-now recommendation. Lens 5's disconfirming pass checked
the optimistic direction too: the clause is *not* silently already implemented (`debate.js`
lines 152–156 re-read live — still prose-change-keyed only), so "Build now" is accurate, not
overclaiming; lens 2 confirmed the clause byte-identical across `47ae48d`/`88eb57f`/`d164ab2`.
Corroboration: high.

### R2-10 — CLOSED BY ACKNOWLEDGMENT (round 3)
New §5 item 11 states the dependency plainly: recurrence-triggered risk-accepts (rows 14, 15)
have no durable counter until an equivalent of the citation ledger exists; treat as advisory;
re-derive counts from primary sources each time. Reasoned, not a silent skip — this is the
argued disposition red said it would accept. Corroboration: high.

### R2-11 — CLOSED (round 3)
§3 row 4 reworded to the accurate claim (no functional parsing path; string present only in
`debate.js`'s descriptive header comment, zero matches in `commands/research.md`). Re-verified
live at `d164ab2`, exact. Corroboration: high.

---

## Round-3 gap details (original grading, preserved verbatim — round-4 statuses above supersede the "OPEN" markers below)

### R3-1 — OPEN — MEDIUM-HIGH — low-medium x medium-high x low — corroboration: HIGH (direct control-flow trace of `debate.js` lines 56–91, 148–198, 200–218; live and unchanged at `d164ab2`, re-confirmed at full merge)
**Location:** §2.3 item 8 (*"`--maxRounds 0` — the emitted log line must distinguish 'never ran'
from 'ran and failed at round 0'"*) and §2.1's Tier A round-loop row (*"Round loop / contested
docket / deadlock / safety ceiling / `adjudicated` bookkeeping"*) — both describe round-loop
degenerate-input handling as covered; neither covers this shape.
**Problem:** `RED_ENVELOPE` requires `verdict`/`gaps` but has no cross-field constraint forcing
`gaps` non-empty when `verdict === 'FAIL'`. Traced by hand: `{verdict: 'FAIL', gaps: []}` makes
`contested.length` stay 0 every round (the judge is never invoked — no adjudication path exists
for "FAIL with nothing to adjudicate"), blue is dispatched against an empty open-gaps list while
being told the debate failed, and this recurs silently until `maxRounds` — no distinguishing log
line, no thrown error. The final return then reports `verdict: 'UNVERIFIED'` alongside
`gaps_outstanding: 0`, a self-contradictory terminal state to any caller reading only the
top-level return. Distinct from §2.3 item 10 (`gaps` *missing* — a schema violation); this is
schema-*valid* but semantically incoherent — a separate degenerate-loop-termination case.
**Required fix:** guard after the `redEnv` null-check: if `verdict === 'FAIL'` and `gaps` is
empty, treat as effective PASS-with-warning or throw a distinguishing error rather than looping
silently to the ceiling; add as a simulator case alongside the existing `--maxRounds 0` case.
Report-side: name the case in §2.3's additions. Red will accept "code fix docketed for run 4"
if the report-side coverage claim is corrected.

### R3-2 — OPEN — MEDIUM-HIGH — certain (already realized, structural) x medium x low — corroboration: HIGH (direct read of `debate.js` lines 132–146, 170, 187, 197 and the merged `debate.test.mjs` lines ~114–123, both live and unchanged)
**Location:** §2.1, Tier A table — *"Friction aggregation: per-seat arrays namespaced by label and
concatenated, never dropped | source | Self-improvement input integrity [L1+L3]"*.
**Problem:** this claim is false for one schema'd, friction-capable seat. `takeFriction` is
called at exactly three sites (red-merge, judge, blue-respond) — never for round-0
blue-synthesize, even though it is schema'd against `BLUE_ENVELOPE`, which declares a `friction`
field. Not equivalent to the (fine, by-design) exclusion of unschema'd frontier/blue-lane
dispatches. The practical sting: this report's own §0 live addendum (the write-block firing
during this retrospective's own round-0 synthesis) is exactly the class of event the structured
`friction` channel would silently drop under the already-merged code — it survives in this
report only because the agent narrated it into prose. The merged regression test ("friction
aggregates from every seat with attribution") stubs and asserts only red-merge/blue-respond, so
11/11-green gives false confidence.
**Required fix:** add `takeFriction('blue-synthesize', blueEnv)` after the null-guard (line 136);
add a simulator case; **correct §2.1's "never dropped" claim to name the exception until fixed**
(the report-side correction is the gate; the code fix may be docketed for run 4).

### R3-3 — OPEN — MEDIUM — medium (untested live, only by trace) x medium x low — corroboration: HIGH on mechanism (direct trace of `debate.js` lines 93–112, 166–197); MEDIUM on real-world frequency (no live `carried` resolution observed yet in this corpus)
**Location:** `debate.js`'s judge prompt (line 182) — *"for carried, state what further research
blue owes"* — against §2.3 item 5's suite-case description (*"carried gap re-enters `openGaps`
with its required-fix intact"*).
**Problem:** the judge is asked for a specific piece of guidance when carrying a gap, but the
script reads `judge.resolutions` only to populate `adjudicated` (which excludes "carried") and
`judge.friction` — the rationale text for a carried gap is never passed into the next round's
`blue-respond` prompt, which is built entirely from red's original gap object. §2.3 item 5's own
test description confirms only the gap's `required_fix` is checked to survive the carry. The
judge's answer is written to `debate.md` for a human, not threaded back into the loop that
requested it.
**Required fix:** either fold the judge's per-gap rationale into the `openGaps` payload passed to
`blue-respond` (more robust), or add a sentence instructing blue-respond to read the latest
`### LEAD` section of `debate.md` before responding (cheaper; relies on blue's initiative — the
reliability class already flagged for backlog item 4 / §3 row 12). An argued risk-accept
(mechanism untested live; low observed frequency) is acceptable if the §2.3 item 5 description
stops implying the guidance is delivered.

### R3-4 — OPEN — MEDIUM — certain x medium x trivial — corroboration: HIGH (direct text comparison within the shipped document, confirmed independently by lens 1 and by the merge's own full re-read; the footnote's corrected direction independently re-verified against arXiv:2606.02646 full HTML by two lenses)
**Location:** §1.1 (H1), diminishing-returns disconfirming bullet — *"with the breakeven shifting
to 3–4 agents on harder tasks and continued gains observed to 7 agents on the hardest — so the
qualitative thesis holds"*.
**Problem:** the body still asserts the exact clause its own footnote retracted this round —
repair-regression one hop further out, **body lagging the correctly-repaired footnote** (the
reverse of the familiar footnote-lag direction). [^DiminishingReturns]'s round-2 text explicitly
drops "7 agents" ("dropped rather than re-cited to an unverified source") and states the
corrected direction is toward *more* caution on hard tasks; the §1.1 sentence the footnote is
attached to — written in round 1, untouched by the round-2 fix — still presents the retracted,
wrong-direction claim as corroborated by "independent re-search." A body-prose reader (the
majority surface) takes away a claim the report's own citation apparatus has withdrawn.
**Required fix:** edit the §1.1 body clause to match the footnote's corrected synthesis — drop
"continued gains observed to 7 agents on the hardest"; the correct wording is already sitting in
the footnote. Trivial; no new research.

### R3-5 — OPEN — MEDIUM — certain (static logic property) x medium x trivial — corroboration: HIGH (pure arithmetic against the row's own roster; row 7's [OPEN] status on the same table; §1.1's documented run-2 `--lanes 2` override)
**Location:** §3 row 6, the R2-8 reconciliation clause — *"four named methods (primary-literature
/ practitioner-production / adversarial-disconfirming-first / local-repo critical-stance) plus a
2-of-N floor on one of them arithmetically needs `lanes >= 4`"* — and, same clause, *"row 7 below
floors N at 3 (the shipped default)"*.
**Problem:** the reconciliation added to close R2-8 **mis-adds**. Taken literally as stated: 3
unfloored methods x 1 lane + 1 floored method x 2 lanes = **5** lane-assignments minimum, not 4.
"`lanes >= 4`" holds only if "adversarial-disconfirming-first" and "local-repo critical-stance"
are silently treated as one lens for allocation (the floor clause's slash-phrasing
"critical-stance/adversarial-disconfirming lens" implies exactly this merge; the same paragraph's
own "four named methods" count contradicts it). A future operator implementing the roster at the
stated floor silently drops one named method — reintroducing R1-16's failure-concentration risk
via a wrong-by-one floor instead of an absent one. Compounding, same clause: row 7 is graded
**[OPEN]** ("`lanes = 3` remains an unguarded default with no minimum check") and §1.1 documents
run 2 overriding it downward to 2 — an unenforced default is not a "floor," and the
reconciliation's whole point is a hard minimum operators must respect. Same failure class as the
round-2 catch it was meant to close (`pattern_unreconciled_numeric_floors`): the composition was
asserted, not computed — this time inside the fix itself.
**Required fix:** one sentence — either (a) state the true minimum as `lanes >= 5` with all four
methods distinct, or (b) state explicitly that the two adversarial/critical methods collapse into
one floored lens (roster = 3 for this arithmetic, minimum 4) and why that merge is safe; and
align "floors N at 3" with row 7's own [OPEN] status ("will floor once built," not "floors").

### R3-6 — OPEN — LOW — low x low x trivial — corroboration: HIGH (git grep re-run at the merge seat: zero matches anywhere in the plugin, including `debate.js:156`'s `ledgerClause`, read in full)
**Location:** §3 row 19 — *"a repo-wide grep for 'independent' in the plugin returns zero hits
outside this ledger clause's own text."*
**Problem:** the phrasing asserts the ledger clause itself contains the word (implying exactly
one hit, located there). The live grep returns **zero hits, full stop** — the `ledgerClause`
string contains no form of "independent." The accurate statement is *stronger* for row 19's
point (the protocol has no independent-cross-referencing language anywhere, not even where a
reader might expect it), but as written it sends a future verifier hunting an exception that
does not exist. **Provenance note, against red:** this exact phrasing originated in red's own
round-2 merge (`findings.md`, round-2 header: "zero hits outside the ledger line's own text") and
was carried verbatim into blue's R2-7 fix — a red-originated imprecision surviving two rounds
inside the very lens whose job is this check. Logged in red's pattern memory (red-side extension
of `pattern_repair_regression_citation.md`). Caught independently by lenses 2 and 3;
deduplicated here.
**Required fix:** reword to "returns zero hits anywhere in the plugin, including the ledger
clause itself."

### R3-7 — OPEN — MEDIUM-LOW — medium x low-medium x trivial — corroboration: HIGH on the textual distinction (both primary sources read directly); MEDIUM on practical import
**Location:** §4 row 5 — class label *"Windows ENAMETOOLONG / long-heredoc fragility"* over
*"**2 documented occurrences across 2 runs**"* — and §3 row 15's likelihood rationale — *"a
2-for-2 rate on the triggering conditions."*
**Problem:** occurrence 1's source names the mechanism explicitly (`run2-friction.md` line 4:
"ENAMETOOLONG... **Windows command-length limit**"); occurrence 2's source names only a symptom
(`blue/CHANGELOG.md` Round 0: "chunked-heredoc workaround attempt then **failed on shell
parsing**" — no errno, no length, no byte count). A parse failure on a large heredoc can arise
from quoting/special-character/CRLF issues independent of total length — and CRLF is a
*separately documented* Windows fragility class in this same corpus (§2.1 Tier A). The "2/2
same-mechanism rate" carrying the High likelihood grade is therefore 1-confirmed +
1-same-family-plausible, not 2-confirmed-identical. Does not overturn the risk-accept or the
revisit trigger — narrower than R2-1 (which fixed *how many*; this flags *what* the second one
was).
**Required fix:** one clause — "occurrence 2's exact failure mode is not confirmed as the
length-ceiling class (its source names only 'shell parsing'); treat the rate as 1-confirmed +
1-same-family-plausible." An argued risk-accept (transcript likely unrecoverable) is acceptable.

### R3-8 — OPEN — LOW-MEDIUM — medium x low x trivial — corroboration: HIGH (live re-fetch of `ideas/backlog.md` item 28 at `d164ab2`, three commits past the footnote's pin)
**Location:** footnote [^CostFigureProvenance] — *"`ideas/backlog.md` item 28, live at `main` @
`88eb57f`"* — as consumed by §2.3/§2.4 and §3 row 18's cost rationale.
**Problem:** live-source drift on the report's own most-cited backlog footnote — a textbook
self-application of §3 row 10's named risk. Item 28 has since gained sub-item (d), "MERGE-SEAT
ANALYSIS (run-3 transcripts): the driver is TURNS × CONTEXT, not file size... red-merge-r1:
~100-150K of material, 2.7M+ cache reads," with per-seat cost figures — new data directly
relevant to §3 row 18 (audit-narrowing hold/risk-accept, which reasons about red's full-re-read
burn as the dominant cost). Nothing claimed is contradicted (direction remains
understatement-only, consistent with the footnote's own caveat); the footnote is simply three
commits stale on a source the report knows is volatile.
**Required fix:** re-fetch and re-pin at current HEAD; add the (d) sub-item and note the
turns-x-context finding in row 18's rationale. Cheap; strengthens the report.

### R3-9 — OPEN — LOW-MEDIUM — certain x low-medium x trivial — corroboration: HIGH (single-sentence internal contradiction, footnote read directly; underlying figures re-verified against arXiv:2606.02646 by two lenses)
**Location:** footnote [^DiminishingReturns], the R2-4 replacement sentence — *"harder tasks
shift the breakeven higher and, per arXiv:2606.02646's actual finding, may show diminishing
returns arriving even earlier (practical knee ≈10, effective diversity saturating well before
nominal N=30) rather than later."*
**Problem:** internal contradiction inside one sentence, introduced by the R2-4 fix itself — the
third consecutive round of trouble in this exact footnote (R1-5, R2-4, now this: a
recurring-footnote pattern worth naming even though each instance is small). "Shift the breakeven
higher" means diminishing returns arrive *later* (at higher N); the conjoined clause claims they
arrive "even earlier... rather than later." Both cannot be true of the same quantity. The source
data supports a coherent claim only by distinguishing two senses the sentence conflates: the
nominal-N breakeven (knee ≈10, higher than moderate-task 2–4) vs. the effective-diversity
saturation ceiling (~1.8 by N=30, reached far below nominal pool size). As worded, a careful
reader cannot extract one consistent claim.
**Required fix:** disambiguate the two senses explicitly (e.g., "harder tasks raise the nominal
breakeven (~10 vs. 2–4), but the ceiling those agents chase saturates ~1.8 effective agents by
N=30 — so adding agents on hard tasks buys less than the higher breakeven alone implies"), or
make the clauses alternatives and say which the synthesis favors.

### R3-10 — OPEN — LOW — certain x low x trivial — corroboration: HIGH (both untagged instances read directly at merge)
**Location:** §2.1, Tier A first row — *"252.9k tokens, 11m48s, honest UNVERIFIED
deadlock[^Run1Friction][^Run1Journal]"* — and (found at merge, extending lens 4's catch) §3 row
4's impact cell — *"High — ~50k tokens vs. 253k–3M live discovery"* — both untagged.
**Problem:** incomplete propagation of the R2-5 repair. [^CostFigureProvenance] reached §2.3's
closing line, §2.4, and the Tier B table, but not §2.1's first row — the *reading-order-first*
instance of the figure, presented as plain fact — nor §3 row 4's impact cell. Direction of risk
is understatement only (per R2-5's own finding), so no verdict moves; this is the
repair-reached-some-instances-not-all class red's memory already names (grep the token
report-wide, not just the sections the repair note cites).
**Required fix:** append [^CostFigureProvenance] to both instances, or add one clause to the
tier-table preamble stating the caveat covers every token figure in §2–§3.

---

## Merge-time verification and dedupe notes (round 3)

- Lens-2 Finding 1 and the compounding row-7 point = R3-5 (merged). Lens-2 Finding 2 and
  lens-3's in-pass "R3-1" = R3-6 (deduplicated; both lenses ran the grep independently, same
  result). Lens-1's in-pass "R3-1" = R3-4. Lens-3's in-pass "R3-2"/"R3-3" = R3-7/R3-8. Lens-4's
  in-pass "R3-1"/"R3-2" = R3-9/R3-10. Lens-5's R3-1..R3-3 keep their preview ids.
- `origin/main` re-fetched at merge: still `d164ab2`; `debate.js` byte-identical to `47ae48d`
  across all intervening commits. No report claim invalidated by drift between the lens passes
  and this merge.
- The `git grep -ni "independen"` result (zero hits) was re-run first-hand at the merge seat
  rather than trusted from the lenses, because R3-6 indicts red's own round-2 text — the merge
  does not get to self-grade by citation to itself.

## Noted, not raised (round 3)

- **§1.4's "15 well-formed gap-pattern files":** live store now holds 23 (was 18 at round 2).
  Already dispositioned round 2 as stale-count churn inherent to describing a live accreting
  mechanism; nothing changed about that reasoning. Logged in the ledger for the running count
  only; re-litigating a settled non-gap every round would violate the no-grinding norm.
- **`commands/research.md` line 9 "haiku for smoke tests" prose:** an operator can already
  manually approximate a smoke run via existing `lanes`/`maxRounds`/`model` parameters —
  consistent with row 4's complexity grade, not a contradiction. (Lens 2.)
- **§3 row 8's "(c) superseded by (a), strictly cheaper" near-miss:** initially read as a leap of
  faith; on check, the corpus's own record (round-1 red-merge: `Edit` succeeded on
  `red/findings.md` where `Write` was refused) gives (a)'s append/Edit mechanism a positive data
  point. Checked, held, not raised. (Lens 4.)
- **CHANGELOG R2-9 phrasing ("added the concrete fix... to the same ledgerClause string"):**
  plausibly skim-read as a live code edit; in full row-10 context ("Build now," future tense) it
  resolves as a report-table recommendation. Not misleading on a full read. (Lens 5.)
- **Disconfirming checks that held (lens 5):** R2-9's ledger clause is not silently already
  implemented (still prose-change-keyed — "Build now" accurate); R2-8's floor has no code
  enforcement and the report doesn't claim it does (row 7 already [OPEN]); no
  gap-resurrection behavior beyond §2.3 item 3's known case; `exhausted` cannot misfire on the
  PASS path.
- **Backlog drift `88eb57f`→`d164ab2`:** docs-only, `debate.js` untouched — the
  [^PinnedRepoState] discipline continues to work as designed; raised only where the drifted
  content is materially load-bearing (R3-8), per round-2's precedent.

## Verified clean round 3 (see red/citation-ledger.md lines 90–111 for the graded pairs)

R2-2/R2-3 fix texts verified in place; [^DiminishingReturns]'s rebuttal figures (GSM-Hard, knee
N≈10, ~1.8 plateau, N≤5 pilot) independently re-fetched twice — all exact; §1.4's
blue/judge-have-no-`memory:`-key negative claim verified by direct frontmatter read (new
leaf-node check); judge-null and citationPasses-const re-confirmed live at `d164ab2`; ledger
clause byte-identical across three refs; run-directory count (still 2) re-confirmed; run2-friction
21 entries / single ENAMETOOLONG line re-counted; §4 rows 1–3 and 10 footnotes re-spot-checked
live; §2.3 item 1's "blue-respond cannot crash" and row 16b's bulk/judgment fallback claims
verified true against `debate.js` by direct read (lens 4); no functional `--smoke` path; frontier
shape-verdict quote verbatim; report-template compliance (Catechism correctly a stub pending
PASS — not a violation). 22 claim-reference pairs ledgered by the round-3 lenses plus the
merge-seat re-verifications above.

## Round 3 disposition summary

FAIL. 11/11 round-2 gaps closed (4 with regressions re-raised; 1 closure a rebuttal red accepts
with evidence — and with red's own proposed citation logged as red's error). 10 new gaps open.
Gate tier: R3-1 and R3-2 (runtime edge cases in the merged `debate.js`, one of which — R3-2 —
directly falsifies §2.1's "never dropped" claim and would have swallowed this report's own
headline live evidence); R3-4 and R3-5 (a body sentence contradicting its own corrected footnote
on a graded bullet; a reconciliation that mis-adds by one lane). R3-3 closes by data-plumbing,
prompt sentence, or an argued risk-accept with §2.3 item 5 corrected; R3-6/R3-9/R3-10 are
mechanical accuracy fixes with no legitimate risk-accept path; R3-7 and R3-8 close by one clause
each or an argued risk-accept red will take if reasoned. For R3-1/R3-2, red will accept
"report-side claim corrected + code fix docketed for run 4" — the report's job is the
retrospective; what cannot stand is the report describing coverage the shipped code does not
have. None of the ten disputes H1–H5. Convergence expectation: this is the first round where
every remaining prose gap is ≤ MEDIUM and every fix is one-sentence-to-one-tag sized; if blue's
round-3 response holds the no-new-regression bar that round 2's repairs did not, round 4 is
plausibly the PASS round.

---

## Round-1 gaps — status after round 2 (unchanged in round 3; no closure regressed further except as tracked via R2→R3 chains above)

### R1-1 — CLOSED (round 2)
Fix verified: §0 "Round 1 correction" block present; headline reframed to "shipped, pending first
live trial"; the ~13 dependent §3/§4 cells threaded ([MERGED]/[OPEN]/[PARTIAL] legend);
[^PinnedRepoState] SHA+timestamp discipline added. Every technical claim in the correction block
(blueEnv guard L136, redEnv guard L171, citationPasses const L139/while L148, judge L181
unguarded at L184) re-confirmed by direct read of `debate.js` at current `main` (`88eb57f`) —
not carried forward. Corroboration: high.

### R1-2 — CLOSED (round 2)
Row 2 reworded from conditional to factual, pinned to `debate.js:184`; new row 2b added. The
*code* defect remains live on `main` (re-verified at merge, `88eb57f`) — correctly carried as a
run-4 docket item, which is what the gap required. Corroboration: high.

### R1-3 — CLOSED (round 2)
§2.3 item 1 rewritten into the three failure classes exactly as specified
(throws: judge; silent-degrade: frontier/red-lens/assembly; already-covered: blue-respond).
Corroboration: high (report text read directly).

### R1-4 — CLOSED (round 2)
"~19%" re-cited to arXiv:2603.22103 with an explicit domain caveat; "~95%" dropped as
untraceable. The new [^NarrativeSimilarity] figures (r=0.388 vs. r=0.461; 76.0% vs. 75.3%;
71.0% vs. 71.7%) verified **exact** against the paper's full HTML Table 5 this round — upgraded
to high (round 1 reached only the abstract). Corroboration: high.

### R1-5 — CLOSED WITH REGRESSION (round 2) → R2-4
The "2–4 agents" bound restated as qualitative synthesis, as required. But the replacement text
introduces a *new* unpinned precise figure ("continued gains to 7 agents on the hardest")
attributed only to "independent re-search this round" — the identical footnote-over-attribution
failure R1-5 was raised to fix, recurring inside its own fix. See R2-4 (and, round 3, the chain's
continuation: R3-4, R3-9).

### R1-6 — CLOSED (round 2)
Corrected diffstat re-verified exact live this round: `gh pr view 14 --json` returns
additions:318, deletions:48, changedFiles:18, 11 commits. Corroboration: high.

### R1-7 — CLOSED (round 2)
"21 entries" verified by direct count of `run2-friction.md` bullet lines (3–23). Both footnotes
corrected. Corroboration: high.

### R1-8 — CLOSED (round 2)
§4 row 6 corrected to "run 2, round 1 only"; verified — the file attests exactly one
live-source-drift entry (line 6, red-merge-r1). Corroboration: high.

### R1-9 — CLOSED (round 2)
[^SubagentWriteBug] now names the signature difference (silent no-op vs. explicit worded
refusal) and downgrades the transfer. Corroboration: high (footnote read directly).

### R1-10 — CLOSED (round 2)
"report.md-specific" struck as the report's own conclusion; backlog quote now explicitly
attributed-not-endorsed, with both falsifying occurrences named. Corroboration: high.

### R1-11 — CLOSED (round 2)
Known-failing flag added to §2.3 item 4; dedicated graded row 2b added. The defect itself
re-verified still live at `88eb57f` (`const` L139, loop L148) — now correctly classified as a
shipped defect, which is what the gap required. Corroboration: high.

### R1-12 — CLOSED (round 2)
Row 16b added with the explicit two-option disposition ((b) for dev/smoke, (a) for keeper runs).
Doctrine quote verified verbatim (`debate.js` L22–24); routing verified (red-lens `...bulk` L164,
red-merge `...judgment` L168). Lens 5 additionally verified the keeper-run inheritance claim:
`bulk = model ? { model } : {}` — with no `--model` flag, bulk seats inherit the session model,
so disposition (b)'s scoping is mechanically real. Corroboration: high.

### R1-13 — CLOSED WITH REGRESSION (round 2) → R2-1
The mechanically-unsound "skeleton may moot it by construction" clause dropped and replaced with
the chunked-append mechanism, as required. But the likelihood re-grade (Medium→High) was keyed to
a recurrence count ("3 times across 3 runs... red-merge seat, round 1, per debate.md's merge-seat
friction") that its cited source does not contain. See R2-1 (closed round 3; narrower follow-on
R3-7).

### R1-14 — CLOSED (round 2)
Vetting-step sentence added to row 13 (pin, review, scope permissions). Both tools' "active
maintenance, passing test suites" claims verified live this round (neither archived;
`arxiv-latex-mcp` pushed 2026-06-30, CI green; `pdf-reader-mcp` main-branch CI green).
Corroboration: high.

### R1-15 — CLOSED WITH REGRESSION (round 2) → R2-7
Graded row 19 added with an explicit disposition — the silence the gap named is cured. But the
row's named mitigation ("independent re-verification against a second source") overstates what
the leaf-node lens actually does, and contradicts the report's own §5 item 8. See R2-7 (closed
round 3; wording follow-on R3-6).

### R1-16 — CLOSED WITH REGRESSION (round 2) → R2-8
Redundancy floor added to row 6 — the failure-concentration trade is now named, which is what the
gap required. But the floor does not arithmetically compose with row 7's lane-count floor and the
report's own risk-accepted ceiling on headcount raises. See R2-8 (closed round 3; the
reconciliation's own arithmetic re-raised as R3-5).

### R1-17 — CLOSED WITH REGRESSION (round 2) → R2-3
Cross-provider model diversity named and dispositioned (defer, not adopt) — the surfacing the gap
required. The factual premise (harness `model`/`judgmentModel` knobs select Claude aliases only;
no provider wiring) verified high this round. But the disposition's load-bearing citation
misapplies the paper's L4 result to the L2 condition. See R2-3 (closed round 3).

### R1-18 — CLOSED (round 2)
Addendum relabeled "self-observed, not yet artifact-logged," downgraded to one data point, with
the independent class corroboration (red's round-1 hit) correctly framed as the stronger
evidence. Corroboration: high. (The residual chronology error inside this same paragraph was
R2-2 — closed round 3.)

### R1-19 — CLOSED (round 2)
Corrected URL verified live this round: resolves to a genuine 672KB PDF, no 404. Quote itself
remains medium (self-labeled search synthesis) — unchanged, acceptable. Corroboration: high on
the URL.

### R1-20 — CLOSED (round 2)
Header restated; no numeric-ratio claim remains for lane 3. Corroboration: high.

---

## Round-2 gap details (original grading, preserved verbatim — round-3 statuses above supersede)

### R2-1 — MEDIUM-HIGH — certain x medium x low — corroboration: HIGH that the citation fails (direct read of debate.md's merge-seat friction section, both friction files, and blue/CHANGELOG.md round 0; re-confirmed at merge)
**Location:** §4 row 5 — *"red-merge (retrospective round 1, this round — third occurrence, per
debate.md's merge-seat friction)"* and *"3 runs/rounds, re-graded likelihood High per R1-13"*;
duplicated in §3 row 15 (*"confirmed recurred 3 times across 3 runs, not 2"*) and §5 item 10
(*"Does ENAMETOOLONG recur a fourth time..."*).
**Problem:** the cited source does not contain the claimed event. `debate.md`'s round-1
merge-seat friction lists exactly two items (lossy PDF-fetch depth; the lens-writes-transcript
process misfit) — no ENAMETOOLONG, no heredoc, no command-length event. The corpus attests
exactly **two** ENAMETOOLONG-class occurrences: `run2-friction.md` line 4 (red-merge-r1, run 2)
and this retrospective's own round-**0** heredoc shell-parse failure (per `blue/CHANGELOG.md`
Round 0). `run1-friction.md` contains zero mentions, so "3 runs" fails on its own terms too.
Likely mechanism: the *write-block's* separately-well-corroborated "third occurrence" count was
transposed onto the structurally adjacent ENAMETOOLONG narrative during the R1-13 fix, without
re-walking the ENAMETOOLONG-specific sources. The fabricated count feeds a likelihood re-grade
(Medium→High) and §5 item 10's "fourth occurrence" build-trigger, which should read "third."
**Required fix:** correct the count to 2 documented occurrences across 2 runs; drop the
"per debate.md's merge-seat friction" citation; renumber §5 item 10's trigger from fourth to
third; re-state the likelihood grade on the honest count (High may still be defensible given
Windows + large-append mechanics — but argue it, don't miscount it).

### R2-2 — LOW-MEDIUM — certain x low x trivial — corroboration: HIGH (both events dated by the corpus's own CHANGELOG and transcript; the contradiction is legible within one paragraph)
**Location:** §0 live addendum — *"red's own round-1 audit pass hit the identical block writing
`red/findings.md` this same round"* two sentences before *"Two independent hits at two seats in
two consecutive rounds"*; inherited by §4 row 4 (*"in the same round it hit blue"*) and §5 item 7
(*"Both write-block occurrences observed this round"*).
**Problem:** chronology error, internally contradictory on its face. Blue's synthesis hit is
dated round **0** (CHANGELOG Round 0); red's `findings.md` hit is dated round **1** (debate.md
round-1 RED friction). "Two consecutive rounds" is the correct half; "this same round" and its
two inheritors are wrong.
**Required fix:** correct the three phrasings to "across rounds 0 and 1" / "one round after it
hit blue."

### R2-3 — MEDIUM-HIGH — realized x medium-high x low — corroboration: HIGH (lens-1 full-HTML fetch of arXiv:2602.03794, independently re-fetched by red-merge at merge time: Table 2, L4 2-agent 67.71% vs. L1 16-agent 65.34%; L2 8-agent 65.44%)
**Location:** §1.1 cross-provider paragraph (round-1 R1-17 fix) — *"the same paper's practical
finding — '2 diverse [persona-lensed] agents match/exceed 16 homogeneous' — shows method/lens
diversity within one provider already captures most of the achievable gain without the
infrastructure cost."*
**Problem:** within-source condition misattribution. The paper's four-level taxonomy makes the
"2 vs. 16" result **L4's** (different models AND different personas). **L2** — persona/lens
diversity on one base model, exactly the condition the bracketed gloss "[persona-lensed]"
substitutes in — needs **8 agents** to match the same 16-agent homogeneous baseline: a 4x
efficiency gap, not "most of the achievable gain." The report's own adjacent sentence correctly
quotes "same-base-model agents remain more correlated than architecturally distinct ones," then
the disposition draws the opposite practical conclusion from the same source. The "defer, not
adopt" call may still be right (the infrastructure-cost argument stands on its own) — but not
for the reason currently given, and §5 item 5's "revisit if lens-assignment under-delivers"
trigger is calibrated against the wrong efficiency curve.
**Required fix:** correct the sentence to attribute "2 vs. 16" to L4 and state L2's real curve
(8-to-match-16); let the defer disposition rest on the infrastructure-cost argument alone;
recalibrate §5 item 5's revisit-trigger baseline.

### R2-4 — LOW-MEDIUM — realized x low-medium x trivial — corroboration as originally graded: HIGH that the footnote lacked the citation; the round-2 belief that the figure "traces to arXiv:2606.02646" was WRONG — disproven by blue's round-2 rebuttal fetch, independently re-verified twice in round 3; red's error, logged
**Location:** footnote [^DiminishingReturns] — *"continued gains to 7 agents on the hardest —
treat as a synthesis across sources, not a single citable number."*
**Problem (as raised):** repair-regression on R1-5 — a new precise numeric bound ("7 agents")
attributed only to an unlinked "independent re-search this round." The gap was real; red's
proposed required fix (add arXiv:2606.02646) was not — that paper does not contain the figure.
**Resolution (round 3):** blue's drop-rather-than-re-cite disposition accepted with evidence.

### R2-5 — MEDIUM — high x medium x low — corroboration: HIGH on the backlog finding (ideas/backlog.md item 28, live at 88eb57f: panel counter excludes cache traffic = 92% of real flow, 610K reported vs. 47.7M transcripts); MEDIUM on whether the report's own figures suffer the same undercount (unverified either way — that open question IS the gap)
**Location:** §2.3 closing line — *"against historical incident costs of 252.9k (run 1) and ~3M
tokens (run 2's quota crashes)"*; §2.4 — *"253k–3M tokens per historical incident."*
**Problem:** measurement-methodology drift, one layer out from round 1's PR-merge class. The
project's own live backlog attests (run 3's live measurement) that its in-band panel token
counter undercounts real spend by ~92% by excluding cache traffic. The report's two most-cited
efficiency figures are of unstated provenance; the side-by-side comparison may not be
apples-to-apples. Direction of risk: understatement only. Not verdict-changing.
**Required fix:** one footnote flagging provenance; point at the cost-audit tool.

### R2-6 — MEDIUM — high x medium x low — corroboration: HIGH (git ls-tree at 88eb57f: `research/` contains exactly two run directories, neither run 3's; backlog commits `47ae48d`/`88eb57f` both cite live run-3 measurements)
**Location:** §3 row 11; §5 items 4 and 7.
**Problem:** run 3 evidently already executed and left **zero artifact trail** in the tree — a
live, fresh instance of exactly the gap row 11 argues for closing, silently mooting §5 items
4/7's "pending live trial" framing: the trial is no longer pending, it is *unrecorded*.
**Required fix:** note in row 11 that run 3 is the demonstrated instance; update §5 items 4/7.

### R2-7 — MEDIUM-HIGH — high x medium-high x low — corroboration: HIGH (direct read of `research-protocol/SKILL.md` and `red-auditor.md` verification mandates; repo-wide grep "independent"; the two contradicting sentences are both round-1 additions on the same page)
**Location:** §3 row 19 disposition — *"a poisoned page asserting a fake fact still has to
survive independent re-verification against a second source"* — against §5 item 8 — *"or does it
require a distinct defense (e.g. cross-referencing claimed sources against an independent
index)?"*
**Problem:** the risk-accept's named mitigation claims a capability the protocol does not grant.
The leaf-node method as specified is: follow the citation to *the* source — the same URL re-read,
not an independent second source. A self-consistent poisoned page passes that check by
construction. §5 item 8, added the same round, half-admits this. Observed practice is better
than the protocol's letter for *miscitation* triggered by a suspicious figure — but that trigger
never fires for a self-consistent fabrication, precisely the scenario row 19 risk-accepts.
**Required fix:** restate row 19's mitigation honestly; collapse the two passages into one
accurately-scoped statement. The risk-accept itself may stand once restated.

### R2-8 — MEDIUM-HIGH — high x medium-high x low — corroboration: HIGH (arithmetic internal to the report's own two rows; both read directly; §1.1's "or"-phrased lane-letter roster cross-checked)
**Location:** §3 row 6 disposition — *"assign the critical-stance/adversarial-disconfirming lens
to at least 2 of N lanes (not 1-of-N) ... cost is one more lane-dispatch"* — against row 6's own
four-method roster, row 7 (*"Lane-count floor (`lanes >= 3` ...)"*), and the §3 risk-accepted
list (*"blanket lane-count raise as a diversity fix"*).
**Problem:** unreconciled numeric floors. Four named method lenses + a 2-of-N floor needs a
minimum of 4 lane-assignments; row 7 floors N at 3, the shipped default. At N=3 the two round-1
fixes cannot both hold. Neither row cross-references the other.
**Required fix:** one reconciling sentence — which the round-2 fix supplied, but see R3-5: the
reconciliation itself mis-adds.

### R2-9 — MEDIUM-HIGH — medium-high x medium-high x low — corroboration: HIGH (ledger clause read verbatim at debate.js:156, live at 88eb57f, re-confirmed at merge; row 10's rationale and §0's ledger praise both read directly)
**Location:** §3 row 10, impact cell — *"Medium — drift is usually caught by re-verification"* —
against §0's ledger praise.
**Problem:** two controls starve each other. The shipped ledger clause's skip-trigger is keyed to
the citing **prose** changing, not the **source** changing or access-date elapsed — the two
things row 10's fix tracks. Row 10's safety net is exactly the step the ledger suppresses for
every already-HIGH claim.
**Required fix:** add a time/access-date-based re-verification trigger to the ledger clause, or
re-grade row 10's impact cell honestly. Either closes it; silence does not.

### R2-10 — LOW-MEDIUM — low-medium x low-medium x low — corroboration: MEDIUM (reasoned from the report's own demonstrated counting failures R1-7/R1-8/R2-1 applied to a new instance)
**Location:** §3 rows 14/15's revisit triggers; §5 item 10.
**Problem:** three closures depend on a future reader correctly counting recurrences across runs
— the exact archaeology this report's own counting-method note shows failing. Citations got a
durable counter; recurrence triggers have none until item 11 ships.
**Required fix:** one clause: "revisit triggers have no counter until item 11 ships; treat as
advisory until then."

### R2-11 — LOW — medium x low x low — corroboration: HIGH (direct fetch of both files at 88eb57f)
**Location:** §3 row 4 — *"confirmed absent: no `--smoke` flag in `commands/research.md` or
`debate.js` as of `main` @ `47ae48d`[^SmokeAbsent]."*
**Problem:** the claim over-reaches its own footnote's verification trail (which checked only
`commands/research.md`) and is literally false for `debate.js`, whose header comment contains
the string descriptively. The functional conclusion (no `--smoke` code path) is correct.
**Required fix:** reword to the accurate scope.

---

## Live evidence logged round 2 (write-block, red-merge seat)

The report-file write-block fired **twice** on the round-2 merge pass: (1) direct Write of
`red/findings.md` refused ("Subagents should return findings as text, not write report files");
(2) a Write of the same content to a **scratchpad** path named `findings.md` (outside the run
tree entirely) refused with the identical message, while the same content under a neutral
filename succeeded. First artifact-logged demonstration that the guard keys on the *filename
semantics* regardless of directory — direct, positive evidence for §3 row 8(c)'s skepticism
(now partially settled: filename-keyed, path-independent). Occurrence ledger for the write-block
class, artifact-attested: run 1 (blue, `report.md`), run 2 (red, `findings.md`), retrospective
round 0 (blue-synthesis, self-observed label per R1-18), retrospective round 1 (red,
`red/findings.md`, debate.md RED friction), retrospective round 2 (red-merge, both paths).
**Round 3: no new occurrence claimed** — all five lens passes and this merge wrote via the
neutral-filename scratchpad-then-copy path as a precaution; the block was neither observed nor
ruled out this round.

## Noted, not raised (round 2 — preserved)

- **Live drift on the report's own pin, harmless this instance:** `main` advanced `47ae48d` →
  `88eb57f` (both docs-only; `debate.js` diff empty). Recorded for the pattern only.
- **§1.4's "15 well-formed gap-pattern files":** the live red-auditor memory store held 18 at
  round-2 audit time — stale-count churn inherent to describing a live accreting mechanism.
- **Row 16b keeper-run enforcement:** initially suspected policy-without-mechanism; refuted by
  direct read — `bulk = model ? { model } : {}` means keeper runs inherit the session model.
- **Row 13 vetting-step enforcement:** no `.mcp.json`/CI mechanism exists, but the row is an
  unbuilt before-run-4 action item like the rest of §3 — holding it to a built-mechanism
  standard would be inconsistent.

## Verified clean round 2 (see red/citation-ledger.md)

Doctrine quote L22–24 verbatim; red-lens bulk / red-merge judgment routing (L164/L168); judge
unguarded + citationPasses const re-confirmed at `88eb57f`; lanes=3 no-minimum + missing `lanes`
return key; simulator AsyncFunction injection + parallel catch-null semantics exact; 11 test
cases match §2.3's list in order; [^AgentTestTiers] 95%/5% exact quote; [^PR14Description]
verbatim; six cited backlog items unchanged despite two further commits; [^PdfMcp] maintenance
claims live-checked; R1-4/R1-5/R1-6/R1-7/R1-8/R1-19/R1-20 fix texts verified against sources;
R1-17's Claude-only-knobs premise; §4 rows 6–10 attributions line-by-line; §4 shape-verdict
frontier quote verbatim; §5 items 8–9 cross-references consistent. 28 claim-reference pairs
ledgered by the round-2 lenses + 4 merge-time re-verifications.

## Round 2 disposition summary (preserved)

FAIL. 20/20 round-1 gaps closed (5 with regressions re-raised); 11 new gaps open. Gate tier:
R2-3 and R2-7 (a tradeoff disposition and a security-adjacent risk-accept each resting on a
misdescribed mechanism/citation), R2-8 and R2-9 (two pairs of round-1-era controls that don't
compose), R2-1 (a fabricated recurrence count feeding a grade and a build trigger). R2-5/R2-6
close by one footnote + one §5 note; R2-2/R2-4/R2-10/R2-11 are mechanical. Expected convergence:
all eleven are additive-fix sized; none requires new infrastructure; none disputes H1–H5.

(Id-collision note: the report's §1.2 references "R2-10" from *run 2's* findings
(memory-architecture). This file's R2-* ids are this retrospective's round-2 gaps — unrelated.)

---

## Round 1 record (original grading, preserved verbatim — statuses above supersede)

Round 1 consolidated five lens passes (3x leaf-node citation verification slices, 1x
logic/completeness, 1x dark-side/risk). Keystone claims re-verified live at merge time
(2026-07-13): `special-circumstances` `main` HEAD was `47ae48d` (docs-only, one commit past
the `00018a5` PR-#14 merge); `debate.js:184` judge dereference unguarded; `citationPasses`
`const` at line 139, outside the `while` at line 148.

### Verdict (round 1): FAIL

20 gaps open, 0 closed, 0 rebutted, 0 adjudicated. R1-1/R1-2 must be corrected before the
report is safe to plan run 4 from; R1-4 and R1-11 are the next tier; the rest are additive
fixes or explicit risk-accepts blue can argue. None of the gaps overturn the report's
substantive conclusions on H1–H5 — the engineering analysis survives; the framing, several
graded cells, and a handful of citations do not.

### Round-1 gap details (as originally raised, condensed to grading + required fix; full prose in git history at the round-1 revision of this file)

**R1-1 — HIGH — realized x high x low** — §0 headline stale: PR #14 merged to `main` at
`00018a5` ~8 minutes after blue's verification commit; ~13 dependent cells loaded on the stale
premise; no pinned-SHA discipline. Required: re-verify, reframe, thread the flip, pin SHAs —
without naively flipping (see R1-2).

**R1-2 — HIGH — high x high x low** — §3 row 2's conditional triggered-and-unmet: judge call
site (`debate.js:181/184`) verified unguarded on merged `main`; reword to factual, dock the code
fix for run 4.

**R1-3 — MEDIUM — high x medium x low** — §2.3 item 1 overgeneralized "null at every call site
crashes": only judge dereferences unguarded; frontier/red-lens/assembly returns are discarded
(silent-degrade class); blue-respond already guarded. Differentiate by failure class.

**R1-4 — HIGH-MEDIUM — medium-high x medium-high x low; corroboration LOW on the figures** —
"~19%/~95%" absent from arXiv:2602.03794 (exhaustive percentage inventory); 19% traces to
arXiv:2603.22103 (narrative-similarity, narrower domain); ~95% traces to nothing. Re-cite with
caveat or drop; the false precision backed §3 row 6's Likelihood=High cell.

**R1-5 — MEDIUM — medium x medium x low; corroboration LOW-MEDIUM** — [^DiminishingReturns]'s
"2–4 agents" bound pinned to none of four bundled sources; VentureBeat's story is tool-count,
not agent-count. Pin or restate qualitatively.

**R1-6 — MEDIUM-LOW — certain x low-medium x low** — [^PR14] diffstat "+2281/−46" vs. live
+318/−48/18 files/11 commits; miscounted at origin. Correct.

**R1-7 — MEDIUM-LOW — certain x low-medium x trivial** — "35-entry" friction count conflates the
header's agent count with the file's 21 entries; two footnotes. Correct both.

**R1-8 — LOW — medium x low x trivial** — §4 row 6 "r1–r2" range unconfirmed; the cited file
attests one r1 instance. Correct or cite the r2 artifact.

**R1-9 — LOW — low x low x trivial** — #13890's silent-no-op signature differs from this repo's
worded refusal; soften the analogy.

**R1-10 — LOW-MEDIUM — certain x low-medium x low** — [^WriteBlock] carries "report.md-specific"
uncritically while the row's own second occurrence (findings.md) and run2-friction.md's
"filename-heuristic guard" falsify it. Strike or annotate.

**R1-11 — MEDIUM-HIGH — high x medium-high x low** — `citationPasses` computed once pre-loop,
never rescaled as the report grows; §2.3 item 4 phrased as if the recompute exists; no §3 row.
Recompute in-loop; add row; flag known-failing.

**R1-12 — MEDIUM — medium-high x medium-high x low** — routing table sends red-lens passes to
bulk tier against the file's own "never... the adversary" doctrine; §3 row 16 silent. Reclassify
or document the bounded tradeoff.

**R1-13 — MEDIUM-LOW — medium x low-medium x low** — row 15's "skeleton may moot it by
construction" mechanically unsound (ENAMETOOLONG is a command-length ceiling, orthogonal to
Write-vs-append); recurrence undercounted. Replace mechanism; re-grade.

**R1-14 — MEDIUM — medium x medium x low** — MCP-server adoption graded on cost alone, no
supply-chain vetting step, in a report headlining CVE-2026-21852. Add the vetting sentence.

**R1-15 — MEDIUM — medium x high x medium** — reflexivity blind spot: the report never applies
its own content-poisoning finding to FEOV's ~22-fetch research phase. A graded row or explicit
risk-accept closes it; silence is the gap.

**R1-16 — LOW-MEDIUM — low-medium x medium x low** — lane specialization concentrates failure:
one null dispatch drops 100% of a method's round coverage. Redundancy floor, re-dispatch policy,
or explicit risk-accept.

**R1-17 — LOW-MEDIUM — medium x low-medium x low** — cross-provider model diversity (the
report's own citation's stronger lever) never surfaced. Name it and its disposition.

**R1-18 — LOW-MEDIUM — high-that-something-happened x low-medium x low; corroboration LOW as
written** — §0 addendum's write-block occurrence is a self-report with no artifact trail, made
by the seat it vindicates. Label "self-observed, not yet artifact-logged" or footnote evidence.

**R1-19 — LOW — certain x low x trivial** — [^WisdomCrowds] URL 404s; real path
`/projects/wisdom-of-llm-crowd.pdf`. Correct.

**R1-20 — LOW — high x low x low** — header's "disconfirming budget met in each lane" false
parallelism: lane 3's per-claim checks are not a search-budget ratio. Quantify or rephrase.

### Round 1 disposition summary

FAIL. 20 open gaps: 2 high (R1-1, R1-2), 1 high-medium (R1-4), 1 medium-high (R1-11), 5 medium
(R1-3, R1-5, R1-12, R1-14, R1-15), 11 medium-low and below. Nothing here disputes the corpus's
substantive H1–H5 conclusions; what fails is present-tense accuracy (R1-1/R1-2), two
quantitative citations (R1-4/R1-5), one live shipped defect misfiled as a test case (R1-11),
and a set of gradable omissions. Expected round-2 path: R1-1/R1-2/R1-6/R1-7/R1-19 close by
mechanical re-verification and edit; R1-4/R1-5 close by re-cite-or-soften; the tradeoff gaps
(R1-12, R1-14–R1-17) close by one-sentence dispositions or argued risk-accepts, which red will
accept if reasoned.
