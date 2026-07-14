# Red round-5 lens 2 — leaf-node citation verification, slice 2 of 3 (§2 Testing strategy, §3 What should change)

Scope: this instance took §2 (`## 2. Testing strategy...` through its four subsections, report lines
481–689) and §3 (`## 3. What should change before run 4...`, lines 692–741) of `blue/report.md`, per
the even three-way split of the report's section headings (instance 1: §0–§1; instance 2 [this
file]: §2–§3; instance 3: §4–§5–Footnotes). Read the full living report in context (911 lines, not
the diff), the full `blue/CHANGELOG.md` (all five rounds, 0–4), and the full `red/findings.md`
(1077 lines) before auditing. Consulted `red/citation-ledger.md` and applied the skip-rule: claims
verified HIGH in a prior round and untouched by round 4's CHANGELOG entries were not re-fetched.
Live-refetched everything round 4's CHANGELOG names as touching this slice (§2.1 gap-id-rollover
row + Tier C bullet; §2.3 additions 13/15; §3 rows 6/13/20/23; the risk-accepted closing paragraph),
because those are this round's actual new surface — everything else in §2/§3 is reused at its
last-recorded ledger confidence.

## Live-repo state at audit time

`git fetch origin main` → HEAD is still `42dba2d` — no drift since round 4's pin. `git log -1
--format="%ci"` on `d164ab2`/`42dba2d` confirms the report's "25 minutes" gap (00:24:02 → 00:49:14)
exactly. `debate.js` re-read in full at `42dba2d`: `RED_ENVELOPE` schema (lines 56–91, no
`supersedes` field), `contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))` (line 178),
`prevGapIds = gapIds` (line 190, full replacement not union — confirms the code-side fix is still
genuinely undocked, consistent with the report's "docketed for run 4" framing), no
`{verdict:'FAIL', gaps:[]}` guard anywhere (consistent with R4-2's decision being report-side only
so far). `ideas/backlog.md`'s docket-detector item re-read via `git show 42dba2d` directly — matches
the report's quoted fix shape (`supersedes: [prior-ids]`, lineage-following at depth ≥ 2, simulator
case) verbatim. Both PDF-MCP repos (`arxiv-latex-mcp`, `pdf-reader-mcp`) re-checked live —
`archived: false` on both, no drift from row 13's claim.

## Finding 1 (NEW, round 5) — §3 row 23 ships a different, less accurate chain-enumeration than the one blue itself adopted for the same finding two sections earlier

**Location:** §3 row 23 — *"this exact corpus contains four live regression chains
(R1-5→R2-4→R3-4/R3-9; **R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8**)"* — against §2.1's Tier A
gap-id-rollover row, sub-row (b), same R4-1 finding: *"this very corpus's `red/findings.md` chain
**R1-5 → R2-4 → R3-4/R3-9**... plus at least three more same-shaped chains... **R2-5 → R3-10**...
**R2-7 → R3-6**... and **R2-8 → R3-5 → R4-3**."*

**Problem:** these are two different lists for the identical claim ("this corpus contains four live
regression chains"), and only the shared first chain (R1-5→R2-4→R3-4/R3-9) matches. The three
chains named in §3 row 23 do not match `red/findings.md`'s own closure record, checked directly:

- `red/findings.md` states **"R2-5 — CLOSED WITH REGRESSION (round 3) → R3-10"** — not R3-8. R3-8
  is a separate round-3 gap (live-drift staleness on the same `[^CostFigureProvenance]` footnote's
  pin, not a content regression minted from closing R2-5) — findings.md's own R3-8 write-up never
  calls it a regression-closure successor of R2-5.
- `red/findings.md` states **"R2-1 — CLOSED (round 3)"** (cleanly, not "closed with regression") and
  explicitly: *"A narrower, distinct follow-on... is R3-7, **not a reopening**."* So R1-13→R2-1→R3-7
  is not a regression-chain in the same sense the row's own claim requires.
- The verified chain for the lane-diversity floor topic is **R2-8→R3-5→R4-3** (confirmed in §2.1
  and in `findings.md`'s own R4-1 gap text), not "R1-16→R2-8→R3-5" — R1-16 (round 1, the original
  redundancy-floor addition) is not recorded anywhere as a "closed WITH REGRESSION" predecessor of
  R2-8, and the row-23 version also silently drops the chain's live third link (R4-3) that §2.1
  includes.

This is not an independent, equally-plausible reading — `debate.md`'s own round-4 BLUE section says
so directly: *"One correction to my own first pass: I initially reconstructed the 'three more
chains' independently... (R1-13→R2-1→R3-7, R1-16→R2-8→R3-5, R2-5→R3-8) before checking [red's]
round-4 RED section, which had already enumerated a different (and, on inspection, more precisely
reasoned) set... Re-verified red's list against the cited rows and adopted it in place of my own"*
(`debate.md`, round-4 BLUE, item 1). Blue explicitly discarded this exact three-chain list in favor
of the one now in §2.1 — but the discarded list is still live in §3 row 23, the row created for the
very same R4-1 gap in the very same round. This is the report's own named
repair-reaches-one-location-not-all class (the pattern behind R3-4, R3-10, R4-4), now occurring
inside R4-1's own fix, between two sections of the *same* round-4 patch.

**Corroboration confidence: HIGH** — pure internal cross-reference (both quotes verbatim from the
current `blue/report.md`), independently confirmed against `red/findings.md`'s own closure records
for R2-1, R2-5, R2-8, R3-5, R3-7, R3-8, R3-10 (all read directly, this pass) and against `debate.md`'s
round-4 BLUE section's own admission of which list is the discarded draft.

**Grading:** likelihood CERTAIN (static text, directly confirmed, not probabilistic) × impact
LOW-MEDIUM (does not change row 23's disposition, grade, or required fix — the count of "four
chains" is right either way, and the fix, adding a `supersedes` field, does not depend on which
specific examples are cited — but three of the four illustrative examples are simply wrong, and a
reader cross-referencing §2.1 against §3 for "the four chains" gets contradictory primary evidence
inside a report whose central argument at this exact location is that citation-lineage precision
matters) × complexity TRIVIAL (replace the three wrong chain labels in row 23 with the ones already
sitting, correct, in §2.1 two sections earlier — no new research, no new verification, a copy-paste
fix).

**Required fix:** edit §3 row 23's parenthetical to `(R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6;
R2-8→R3-5→R4-3)`, matching §2.1 and `findings.md`'s own R4-1 record exactly.

## Finding 2 (NEW, round 5, minor) — §2.3 addition 15's worked example labels the R1-5 chain's second link "WITH REGRESSION," but that link's actual closure status is "rebuttal accepted with evidence"

**Location:** §2.3, addition 15 — *"three canned `redEnv` round objects where round 1 raises gap
`X-1`, round 2's merge closes `X-1` 'WITH REGRESSION' and raises a fresh-id successor `X-2`... and
round 3's merge closes `X-2` 'WITH REGRESSION' and raises `X-3`/`X-3b`"* — mirroring the real chain
R1-5→R2-4→R3-4/R3-9.

**Problem:** `red/findings.md` records R1-5's closure as *"R1-5 — CLOSED WITH REGRESSION (round 2)
→ R2-4"* — matching the addition's first link. But R2-4's own closure is recorded as *"R2-4 —
CLOSED, **REBUTTAL ACCEPTED WITH EVIDENCE** (round 3) → regressions R3-4, R3-9"* — a different
closure-status label than "WITH REGRESSION," even though it still mints two fresh-id successors
addressing the same underlying defect. Addition 15's abstraction applies the "WITH REGRESSION" label
uniformly to both links, which is accurate for link 1 but not for link 2 as this corpus's own record
actually shows it. The underlying code behavior the simulator case needs to test (does `supersedes`
get set, and does the docket arm by round 3) does not depend on which closure-status label produced
the successor id — so this does not change the case's validity as a test — but the case's own
framing claims to "mirror this corpus's own chain directly," and as worded it mirrors a
simplified, not-quite-accurate version of it.

**Corroboration confidence: HIGH** — direct comparison of the addition-15 text against
`red/findings.md`'s own R1-5 and R2-4 closure-status lines, both read verbatim this pass.

**Grading:** likelihood CERTAIN (static text) × impact LOW (does not affect the test's validity or
the required code fix; a documentation-fidelity nit on a proposed, not-yet-built test case) ×
complexity TRIVIAL (either loosen the wording to "closes ... and mints a fresh-id successor
(whatever the closure status)" or split the addition into two sentences naming both real
closure-status labels).

**Required fix:** optional / low priority — reword to avoid implying every regression-chain link is
closed under the identical "WITH REGRESSION" status, since this corpus's own worked example has one
link closed that way and one link closed via rebuttal-acceptance.

## Verified clean this round (added to `red/citation-ledger.md`)

- `debate.js`'s `RED_ENVELOPE` schema (no `supersedes` field, lines 56–91), `contested` filter (pure
  id-equality, line 178), `prevGapIds = gapIds` (full replacement, line 190, not union — code fix
  still undocked) — direct read at `origin/main` @ `42dba2d`, this round. HIGH.
- `ideas/backlog.md`'s docket-detector item, `git show 42dba2d` — quoted fix shape
  (`supersedes: [prior-ids]`; lineage-following at depth ≥ 2; simulator case) matches §2.1's and
  §3 row 23's citation of it verbatim. HIGH.
- Commit-timestamp delta `d164ab2`→`42dba2d` = 25m12s, confirming the report's "25 minutes" claim
  exactly. HIGH.
- §2.1 Tier C lossy-fetch bullet's `MA-`-prefixed ids (`MA-R1-19, MA-R1-28, MA-R2-8 residual,
  MA-R3-14, MA-R3-15, MA-R4-9`) and §3 row 13's matching set — both re-verified as real gap ids in
  `research/2026-07-12_memory-architecture/red/findings.md` (direct grep, this round: all six exist
  verbatim at the cited severities). Report-wide grep confirms exactly 4 locations carry the `MA-`
  prefix fix (§1.2, §2.1, §3 row 13, §4 row 1 — the last two outside this slice, spot-checked for
  cross-slice consistency) and this retrospective's own bare `R1-19` (WisdomCrowds footnote, a
  different gap, this corpus's own id) is correctly left unprefixed per the stated disambiguation
  rule. HIGH.
- §3 row 6 (R4-3 fix): the operative floor sentence now reads "the adversarial-disconfirming-first
  lens (a distinct method from local-repo critical-stance, named separately below)" — the slash
  ambiguity is gone; no residual "lanes >= 4" assertion found anywhere in the report except inside
  text explicitly narrating the retracted figure (`grep -n "lanes >= 4"` — both hits are
  self-labeled as the wrong number being corrected, not asserted). HIGH.
- §3 row 20 / §2.3 addition 13 (R4-2 decision): the thrown-error message text is identical in both
  locations, verbatim; no guard code exists yet on live `main` (consistent — decided, not shipped).
  HIGH.
- Risk-accepted closing paragraph (R4-4): `grep -n "4th|fourth"` across the whole report — every
  live instance now reads "third, not a fourth" or is a historical quote of the retracted figure
  explicitly marked as retracted (line 81's CHANGELOG-summary mention); zero remaining unflagged
  "pending a 4th occurrence" instances. HIGH.
- Row 13's PDF-extraction tool claims (`[^PdfMcp]`) — live-rechecked this round (`gh api .../repos`,
  both `archived: false`); untouched by round-4 CHANGELOG in substance, reused/reconfirmed. HIGH.
- §2.2, §2.4 (simulator design, three-tier stack) — untouched by round-4 CHANGELOG; spot-read for
  internal consistency against the live code, no new defect found. Reused at prior HIGH.
- §2.3's 11+15 test enumeration — numbering (1, 2, 3, 15, 4–14) intentional and explicitly dated
  ("Added round 4, R4-1") at item 15's insertion point; no duplicate or missing item numbers.

## Summary disposition (this lens only)

2 new gaps in §2–§3 for round 5: one LOW-MEDIUM finding (§3 row 23 carries blue's own discarded,
factually-wrong chain enumeration for R4-1, while §2.1 two sections earlier carries the corrected
one blue explicitly adopted in `debate.md` — a same-round, cross-section propagation miss on the
report's own headline round-4 finding) and one LOW, largely cosmetic finding (addition 15's uniform
"WITH REGRESSION" labeling doesn't match this corpus's own mixed closure-status record for the chain
it claims to mirror directly). No round-4 fix in this slice (R4-1's §2.1/§2.3 content, R4-2, R4-3,
R4-4, R4-5's in-slice locations) is reopened or found defective on the merits — every required fix
landed correctly where the CHANGELOG says it landed; the new findings are both propagation/precision
misses on otherwise-sound repairs, consistent with this corpus's declining-severity trend (round 3
worst: 2 MEDIUM-HIGH; round 4 worst: 1 HIGH (R4-1, a substantive finding, not a repair-of-a-repair);
round 5, this slice: worst LOW-MEDIUM, both are repair-of-a-repair precision misses).
