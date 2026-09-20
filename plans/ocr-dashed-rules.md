# A rule drawn as dashes

Status: revision 3, IMPLEMENTED and measured 2026-09-20. Carries #1027. §VIII holds the result,
including a correction to §IV D5's central number.
Revisions 1 and 2 each failed their audit; §0 says what those verdicts changed. **[r3]** marks
what moved since revision 2.

## 0. Revision 3 — what two failed audits changed, and what #1074 settled

Revision 1 failed on eight gaps; revision 2 failed on four more, the sharpest being that its
central fix repaired a gate this plan's pages never reach. Both verdicts stand and the design
below is different because of them. **#1074 merged 2026-09-20 (PR #1096)** and settles what
revision 2 was blocked on.

**What #1074 changed for this plan, read off the code rather than assumed. [r3.1]** Revision 3
said the repaired boxes "arrive already normalized". **That is false of every path that exists**,
and three audits did not catch it because none of us had traced the one function that decides:

```
readPage           DetectGrid → if !thr.Table(grid) → PageText, RETURN      ← a dashed page exits
HERE
                   else → readGridPage
readGridPage       TSV, rotation probe, headerBand (page pixels), GridLines
                   ── THE BOUNDARY (page.go:257) ── normalize lattice + words
                   levels, Reconstruct, TextCells
```

A dashed page is exactly a page the detector rejects, so it returns at `readPage`'s `!Table`
branch and **never reaches the only normalization crossing there is.** The repair's call site is
above the boundary and its boxes are in PAGE pixels when it gets them.

**So the boundary moves, and it moves to one owner.** `readPage` acquires the lattice — from the
detector, or from the repair when the detector rejected the page — normalizes it ONCE, and passes
the normalized lattice and the scale into `readGridPage`, which then normalizes only the words.
Today `readGridPage` fetches and normalizes its own lattice; after this it does neither, because a
lattice normalized twice is a lattice scaled by `(300/dpi)²` and nothing would say so.

That restructuring is the load-bearing part of this plan. It is what makes the constant split
below true rather than aspirational:

- **In C, on real pixels:** the closing length, derived for the page's DPI as `SEL` is (`GridFor`).
- **In Go, after `readPage`'s crossing:** the thinness cap, the span requirement and D3's crossing
  test, stated at 300 and never scaled.

Revision 2 treated all three as one kind and stated them in 350-DPI pixels, which was the audit's
finding about units. There is no third kind and no list to keep: the layer a constant acts in
decides how it is written.

**The withdrawn D6, re-aimed.** Revision 1 was told the dropout gate goes dead on a repaired page;
revision 2 fixed that by filling `GridStats` with repaired crossings. Revision 2's audit showed the
fix was aimed at a branch a dashed NUMERIC matrix never reaches: `isStrongMark` needs an x/X token,
so `Reconstruct` returns `ErrNoMarks` and `fallbackReason` returns at its first case. **The gate
that actually publishes or refuses such a page is `CellStats.refuse()`** — `MinTextCellColumns`,
`MinTextCellRows`, `MinWordsPerCell` — and after #1074 those run against normalized geometry, so
they are correct at the dashed page's 350 DPI rather than off by a sixth.

**This plan therefore adds NO new acceptance gate.** A repaired page reaches `TextCells` and is
judged by the same one every ruled text table is judged by. What the record must still carry is a
different duty from gating, and is D6 below.

## I. Summary & Goals

**The defect.** NBS SP 602 rules its tables with typewriter dashes. The detector found **2 of its 60
pages** and missed **all 12 dashed-rule tables** (#934). A dashed rule is not a faint rule: the
morphological opening deletes every run shorter than its own length, so a rule made of 20 px dashes
is **invisible**, not weak. No threshold recovers it — measured at SEL 151, 101, 61 and 31, the
horizontal rules survive none, and the SEL that admits them (15) calls plain prose a table. 600 DPI
reproduces the failure exactly, because the dash-to-gap ratio is scale-invariant.

**Goals.**

1. A table ruled with dashes is detected, and its rules reach the lattice so #932's cells and #933's
   levels work on it unchanged.
2. A page the existing detector already fires on reads **byte-identically** — the repaired path is
   consulted only where the existing one found nothing (§IV D4). **[r2]**
3. On the tuning document, the repaired path adds **exactly the pages named in §IV D5 and no
   others**. It does add one. Revision 1's "the verdicts do not move" contradicted D5; the criterion
   is *the named exception, and no unnamed one*. #1027's acceptance ("TP=33 FP=1 FN=0 … do not
   regress") is amended in the issue before this ships, or this plan does not ship. **[r2]**
4. A page detected by the repaired path carries a rule measurement that is **its own** and says
   which path found it — never the existing path's zero read as a measurement (§IV D6). **[r3: the
   gating half of this goal is withdrawn; §0 says why. What remains is a record duty.]**
5. The repaired verdict is **resolution-invariant**, like every other geometric decision after
   #1074, and §VI.8 is the gate rather than a measurement someone remembers to take. **[r3]**

**Non-goal.** Reconstructing SP 602's dashed tables well. Detection is this plan; what the
reconstruction then does with rules it can finally see is measured after, on the corpus.

**The value case, weighed.** The machinery is real: a new C entry point, one new derived length and
two new 300-space constants, a second lattice source, and a new field on the record. **[r3: it is
smaller than revision 2's, because #1074 took the second acceptance gate out of it — the repaired
page is judged by `CellStats.refuse()` like any other ruled text table.]**
Against it: a 60-page document reads twelve of its tables as prose today, and the class is not
one document — typewriter and teletype rules are dashes by construction, so every scanned technical
report of that era is in it. The naive alternative (ship nothing, accept the residue) was the
status quo through #934 and is what #1027 exists to overturn. The fork is recorded because it is a
real one, not because it is close.

## II. Technical Context

- **Where the code runs.** The detector's pixel work is C (`internal/tessocr/shim.{h,cpp}`) behind
  the `tessocr` build tag; everything above it is Go that must compile on the untagged build, where
  `engine_stub.go` refuses every entry point. A new entry point needs **both halves** or `go test
  ./...` does not build on any leg.
- **Resolution is per page.** #1031 renders each page at its own native DPI, floored 300 and capped
  600 (`fetchcache/pagerender.go`), and the detector's constants are derived from DPI by
  `GridFor(dpi)`. A constant measured at one DPI and stated in pixels is a bug in this file; the
  convention is measured-pixels over measurement-DPI, pinned by
  `TestTheDerivedTuneIsTheMeasuredTuneAt300`.
- **The pixels are normalized before anything reads them.** #1055 put `pixBackgroundNormSimple` in
  `read_png`, so the repair operates on a background-flattened page, not the raw scan.
- **The tuning document cannot be committed.** IEEE 1012 is copyrighted (`references.json`), which
  is why the tune is gated by `testdata/grid300-sel151.txt` — three scalars per page — rather than
  by the pages. Any pin this plan proposes over those 80 pages has to live in a record of that kind.
  §V.2 says what shape it takes.

## III. What the page actually looks like

Measured on `testdata/corpus/nbs602-dashed-matrix` at its native 350 DPI:

| | dash length | gaps |
|---|---|---|
| horizontal rule | 19–21 px | median 4–6, p90 12–22, max 48 |
| vertical rule | 13–15 px | median 4–12, p90 18–19, max 21 |
| **prose on the same page** | ink runs median 7 px | **median 11, p75 20, p90 44** |

**The gaps overlap.** A closing long enough to bridge a rule's dashes also joins the characters of a
word, so "close the gaps" alone turns every line of prose into a rule. That is why the obvious fix
is not the fix, and it is the thing to remember when this is read later.

## IV. Proposed Changes

**D1. Repair, then measure what survived.** Three steps on the binarized page, per axis, all in C
on the page as scanned:

1. **close** by the dash gap — joins a rule's dashes, and unavoidably the characters of a word;
2. **open** by the rule length (the existing `ruleRunInches`, unchanged and not re-coined) — keeps
   long runs, of which a closed line of prose is still one;
3. **keep only components thin in the perpendicular axis** — a rule is thin; a line of prose closed
   into a run is as tall as its x-height.

**The constants, split by the layer they act in (§0). [r3]** Revision 2 wrote all three as
350-DPI pixels. One is morphology on real pixels and belongs with `SEL`; the other two read boxes
that arrive normalized and belong with `ruleJoin`:

| constant | layer | kind | measured | written as |
|---|---|---|---|---|
| `dashGapInches` | C, page pixels | length ∝ DPI | 25 px at 350 | `25.0 / 350`, via `GridFor` |
| `maxRuleThick` | Go, 300-DPI space | length, never scaled | 12 px at 350 = **10 px at 300** | a package constant beside `ruleJoin` |
| `minRepairedSpan` | Go, 300-DPI space | length, never scaled | 600 px at 350 = **514 px at 300** | a package constant beside `ruleJoin` |

**Only the closing length is new in C**, because only it is morphology; the opening reuses `SEL`
unchanged. The other two run AFTER `readPage`'s crossing (§0), on boxes it has just normalized, so
they are 300-DPI numbers and scaling them again would apply the correction twice.

Step 3 is therefore a Go box test rather than a C morphological step — which costs a larger dump
across the boundary and buys the thing §VI.6 needs: a filter that can be deleted in the default
build, on a record (§VI.2) that holds what the openings left rather than what the filters kept.

Measured on the dashed matrix at its native 350: this recovers **8 horizontal rules** at y = 953,
1195, 1386, 1579, 1722, 1914, 2105, 2296 — gaps of 242, 191, 193, 143, 192, 191, 191, the 143 being
the header band, so they are regularly and not evenly spaced — and the vertical rules that bound its
columns, where the current detector recovers nothing at all. **Every geometric number in §IV was
measured before #1074 merged, in page pixels at 350; each is restated here in the space its
consumer now reads, and §VI.1 re-measures them rather than trusting the conversion. [r3]**
**D2. Thinness is not enough, and two cheaper ideas do not work.** Photo-reduced dot-matrix print is
*also* long, thin and uniform. Measured over the corpus, neither candidate discriminator separates a
repaired rule from a repaired line of small print:

| | dashed matrix (a rule) | DOE dot-matrix listing (not a rule) |
|---|---|---|
| original ink coverage along the run | 0.53 | 0.44 — and a contents page reads 0.51 |
| thickness variation along the run | CV 0.10 | CV 0.11 |

So the discriminator cannot be a property of one run. **It has to be the structure the runs form.**

**D3. A repaired table must be a LATTICE.** The repaired path declares a table only when, among
components that are thin AND span at least `minRepairedSpanInches`:

- at least **2 vertical rules each cross at least 2 horizontal rules**, and
- at least **2 horizontal rules are each crossed by at least 2 verticals**.

A table satisfies this by construction. Aligned text does not: a monospace listing's character
columns break at every space, and its lines are crossed by nothing.

Measured over all eight corpus pages — the row this plan exists for is the first **[r2: the three
pages revision 1 omitted are now here, including the blank]**:

| page | long h | long v | verdict | truth |
|---|---|---|---|---|
| `nbs602-dashed-matrix` | 8 | 8 | **TABLE** | table |
| `doe-dotmatrix-listing` | 3 | 4 | no table | not a table |
| `doe-equations-figure` | 0 | 1 | no table | not a table |
| `nbs602-contents` | 0 | 1 | no table | not a table |
| `nbs602-cover-photo` | 0 | 0 | no table | not a table |
| `nbs602-blank` | 0 | 0 | no table | not a table |
| `nbs602-form` | 14 | 0 | no table | table — **found by the existing path** (D4) |
| `usfs-birds-markgrid` | 0 | 0 | no table | table — **found by the existing path** (D4) |

**D4. The repaired path is CONSULTED ONLY WHERE THE EXISTING ONE FOUND NOTHING. [r2]** Revision 1
said "additive" and did not say when the repaired rules reach `Cells`/`Tables`, which left three
readings. The decision: `GridLines` is tried first, unchanged; the repair runs **only if it returns
no table**, and then the repaired rules are the *only* lattice source for that page. There is no
union. A page the existing detector fires on never sees a repaired rule, so its reconstruction
cannot change — the closing merges adjacent rules, the thinness cap drops thick borders, and the
span requirement deletes short verticals that `lattice.go`'s column coverage relies on, and none of
that reaches a page that was already working.

This is also measured necessity, not only caution: the ruled form (`nbs602-form`, 14 long
horizontals, 0 long verticals — a form's boxes are one row tall) and the shaded mark grid
(`usfs-birds-markgrid`, 0 and 0 — its "rules" are shading edges) are both detected by the existing
path and **not** by this one. A replacement would lose them.

**D5. The cost, bounded and named — and its failure direction is UNMEASURED. [r2]** Over the 80
pages of the tuning document, rendered at their native 324 DPI **[r2: revision 1 did not state the
DPI]**, the repaired path fires on 30, of which 29 are pages the tune already calls tables. The one
addition is **p0022** — Figure 2, a landscape diagram whose dotted horizontal guides repair into
rules and whose dashed boxes supply the verticals. It is a figure, not a table.

That takes the tune from 1 standing false positive to 2 (precision 0.971 → 0.946, recall unchanged
at 1.000).

Revision 1 claimed p0022 joins p0025's FIRES-AND-DEGRADES-HONESTLY class. **That claim is withdrawn
as unearned.** p0025 degrades honestly because its box binds no usable cells, which
`CellStats.refuse()` catches; the repaired path fires only where D3 found a *crossing lattice*,
which is the condition `refuse()` tests for. p0022 is therefore more likely to clear it than p0025
ever was, and what clearing it produces is a fabricated `|`-separated table of a figure, quotable
downstream through `cite --ocr-quote`. **§VI.5 is where that is settled, with an accept/reject
fixed before the number is known:** p0022's actual `Text`, `TextCells`, `TextCellFallback` and
`RuleSource` are recorded in this section before merge, and a page that PUBLISHES a table of a
figure blocks the change. **[r3: revision 2 required
only that the output be "recorded", which a fabricated table passes by being written down.]**

**What #1074 changed about the odds here, corrected. [r3.1]** Revision 3 said `MinWordsPerCell` had
been applied a sixth too small. **That is wrong:** `WordsPerCell()` is `WordsPlaced/Cells`, a ratio
of COUNTS, and `normalize.go` states that counts and ratios do not cross the boundary — it was never
scaled at all. What #1074 moved for `refuse()` is which CELLS EXIST, through `ruleJoin`,
`minBandHeight` and `cellPad`, and therefore `Columns` and `Rows`. The conclusion is unchanged and
the mechanism was invented; §VI.5 is the gate either way, which is why it is a gate.

**D6. The record says which path found the rules, and no zero stands for a path. [r3]** This is
what survives of revision 1's D6 after revision 2's audit took the gating half away. The duty left
is a RECORD duty: on a page detected only by the repaired path, `GridStats` is what the detector
measured, which is ~0 — and a receipt reading `grid intersections: 0` beside `table: true` is a
fact nobody can act on, because it reads identically to a page with no rules at all.

So the page result gains `RuleSource`, and the rules it names are the ones the lattice was built
from. It is set on every page that carries a table verdict, its values are `detector` and `repair`,
and **its zero value is the empty string, which denotes NO PAGE-LEVEL VERDICT rather than the
detector** — the inference-from-a-zero revision 2's audit called out. A page with `Table: true` and
an empty `RuleSource` is a bug, and `TestEveryTablePageNamesItsRuleSource` fails on it.

**The repaired crossing count names its field, its space and its basis. [r3.1]** Revision 3 said
only that it "goes on the record", which left three things silent and each of them silent in the
way this plan exists to stop. Stated:

- **Field:** a new `RepairedCrossings`, NOT a reuse of `GridIntersections`, whose own doc comment
  says it is "the detector's measured rule-crossing count AT THE PAGE'S OWN RESOLUTION". Reusing it
  would make the field's documentation false on exactly the pages the field matters for.
- **Space:** the 300-DPI space, stated on the field, because a reader divides it against
  `ExpectedIntersections()` and that denominator is a count of lattice points.
- **Basis:** computed in **Go, over the FILTERED boxes the published lattice was built from** — not
  in C. D8 makes C filter nothing, so a C-measured count would be taken over components the Go
  filters then discard: a number that does not measure the lattice it sits beside, which is the
  defect D6 was re-aimed to remove. **[r3.1: revision 3 had C measure it, and D8 made that wrong.]**

`RuleSource` travels receipt → reading so the corpus golden can print it; `renderGolden`
(`goldens_cgo_test.go`) and the corpus golden line build from different structs and both need it,
or eight goldens read as if one path produced them.

**Where the invariant is enforced. [r3.1]** `Table: true ⇒ RuleSource != ""` is checked where the
receipt is WRITTEN — the `pageReceipt` → `pageReading()` projection — and its test runs in the
DEFAULT build. A unit over hand-built `PageResult` values passes on fixtures it constructs and never
sees a page, and the corpus leg that would see one runs only under `FEOV_OCR_CORPUS=1`, which §VI.9
states is not in `check`. That is the same projection that silently dropped `DPI` in #1058.

**D7. Failure contract, covering the counts as well as the geometry. [r3]** `tessocr_repaired_rules`
follows its siblings: `nullptr` is a failure (decode, binarize, morphology, allocation), an empty
geometry string is a MEASUREMENT — a page with no repaired rules. Revision 2 stopped there, and its
audit found the hole: the counts travel in the same text dump, and `ParseLattice` `continue`s past
any line it cannot parse, so a missing or garbled count line yields `Intersections: 0` — reviving
exactly the silent zero D6 exists to remove.

Therefore **the counts do not travel in the dump.** They are `int` out-parameters beside a status
code, the shape `tessocr_grid_stats` already uses, so a count that did not arrive is an error at the
call and never a zero that reads like a clean page. Because D4 makes the repair a second opinion, a
repair FAILURE would otherwise degrade to "not a table", which is indistinguishable from "no dashed
rules here" on the one class this plan exists for.

**D8. The thin and span filters and the crossing test live in Go.** C returns the closed-and-opened
components as boxes and nothing else; the thinness cap, the span requirement and D3's crossing test
are Go, so they are unit-testable in the default build and each is a deletable row (§VI.6) that does
not need the C stack to kill. **C FILTERS NOTHING.** **[r3: revision 2 left
the thinness cap ambiguous between the two layers, and the audit found what that costs — §VI.2's
record would hold post-filter boxes, so deleting the Go filter could not turn its test red. The
record holds what the two openings left, and every filter that decides a verdict is deletable.]**
## V. Components

**Census — every carrier of a grid fact.** Sweep, re-run for revision 3:
`grep -rn "res\.Table\|\.Grid\b\|GridStats\|GridIntersections\|GridLines\|Lattice\|TextCells(" plugins/frank-exchange-of-views/tools --include=*.go`
It returns 19 files. Revision 2 listed eleven of them; the eight it omitted are marked **[r3]** and
were found by the audit re-running this plan's own command.

- `[MODIFY] internal/tessocr/shim.{h,cpp}` — `tessocr_repaired_rules(png, len, close, sel, ...)`:
  the closed-and-opened components as boxes, UNFILTERED (D8), with the counts as out-parameters
  beside a status code (D7).
- `[MODIFY] internal/tessocr/engine_cgo.go` **and** `internal/tessocr/engine_stub.go` — both halves,
  or the untagged build does not compile on any leg.
- `[MODIFY] internal/tessocr/tessocr.go` — `dashGapInches` joins `GridFor(dpi)`; `maxRuleThick` and
  `minRepairedSpan` are package constants in the 300-DPI space (D1).
- `[MODIFY] internal/tessocr/lattice.go` — D3's crossing test and D8's filters, over repaired boxes.
- `[MODIFY] internal/tessocr/lattice_test.go` — **[r3]** 18 hits, the most of any file in the
  census, and the home of the crossing test's and both filters' unit tests, which §VI.6 requires be
  killable in the default build. Revision 2 asserted four deletable rows with no stated home.
- `[MODIFY] internal/tessocr/page.go` — D4's second-opinion ordering, and `RuleSource` (D6).
- `[MODIFY] internal/tessocr/levels.go` and `levels_test.go` — **[r3]** `LevelBandCells` consumes
  the `Lattice`. §I goal 1 claims #933's levels work unchanged on repaired rules; §VI.5 is where
  that stops being a claim.
- `[MODIFY] internal/fetchcache/pageread.go` — `pageReceipt`/`PageReading` carry the repaired
  measurement and `RuleSource`; the doc comment at 145-151 states the dropout gate's units.
- `[MODIFY] internal/fetchcache/pageread_test.go` — **[r3]** it asserts the receipt→reading
  projection field by field, so a new field lands unasserted unless it is added here. This is the
  class CLAUDE.md names: a field nothing asserts can stop being written while every row still reads
  as complete — and it is exactly the defect I shipped in #1058's projection.
- `[MODIFY] internal/tessocr/goldens_cgo_test.go` and `internal/fetchcache/corpus_cgo_test.go` —
  `renderGolden` prints `grid h=%d v=%d intersections=%d`; `RuleSource` belongs on that line or
  every golden reads as if one path produced it.
- `[MODIFY] internal/tessocr/engine_cgo_test.go` — **[r3]** it pins a `GridStats` literal for
  `DetectGrid`; the tagged test of the new C entry point belongs beside it and revision 2 named no
  home for it.
- `[MODIFY] internal/tessocr/textcells_cgo_test.go` and `internal/cli/fetchocr_test.go` — **[r3]**
  both in the sweep; checked for whether a repaired page changes what they assert.
- `[MODIFY] internal/tessocr/testdata/gen/pages.go` — **[r3]** the generated fixtures. A DASHED
  fixture belongs here: it is the only way the repair gets a page whose truth is known by
  construction rather than transcribed by eye.
- `[MODIFY] internal/tessocr/normalize.go` — **[r3.1]** the crossing moves to `readPage` (§0) and
  this file owns it; the header comment's "one boundary in readGridPage" becomes true again.
- `[MODIFY] internal/tessocr/normalize_test.go` and `internal/tessocr/page_test.go` — **[r3.1]**
  §VI.8's arm, and the pure-Go home of `readPage`'s decision.
- `[MODIFY] internal/tessocr/engine_stub_test.go` — **[r3.1]** `TestStubRefusesLoudly` is a
  HAND-KEPT roster of entry points (New, PageText, PageTSV, RotatedBand, DetectGrid, ReadPage). The
  new one joins it, or a stub returning `(Lattice{}, nil)` compiles, passes every default-build
  test, and reads on every untagged build as "this page has no repaired rules" instead of the named
  refusal `engine_stub.go`'s own comment requires.
- `[MODIFY] internal/tessocr/detector_test.go` — **[r3.1]**
`TestTheDerivedTuneIsTheMeasuredTuneAt300`
  compares `GridFor(300)` by struct equality against a literal, so adding `dashGapInches` to
  `GridThresholds` turns it red until the literal carries it. That is the gate working.
- `[NEW] internal/tessocr/testdata/repaired300.txt` — §VI.3's record, named for the space it is
  stated in. **[r3: revision 2 called it `repaired350.txt` and tagged it `[MODIFY]`.]**
- `[MODIFY] internal/tessocr/detector_test.go` — p0022 joins the boundary set with its **measured**
  outcome (D5), not an asserted class.
- `[MODIFY] testdata/corpus/*` — `nbs602-dashed-matrix` changes verdict; its `expect` already
  demands `table: true`. All eight `reading.golden` files regenerate regardless, because line 1
  carries `DefaultPageEngine.Identity()`, which hashes this package's source — **comments
  included, which is how PR #1096 went red on CI while passing locally. [r3]**


## VI. Verification Plan

Each step in its own shell; the cstack environment leaking into a golden step makes it fail. Run
from `plugins/frank-exchange-of-views/tools`, after
`unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. **Re-measure §IV in the space its consumers now read. [r3]** Every geometric number in this plan
   was measured in page pixels at 350 before #1074 merged. Before anything is implemented, the dash
   gap, the thinness cap and the span are re-measured on `nbs602-dashed-matrix` through the
   normalized path, and §IV's table is corrected from that run rather than from the arithmetic in
   D1. A converted number is a claim about the conversion; this is the measurement.
2. **Default build.** `go build ./... && go vet ./... && go test -count=1 ./...` → **0 failed**.
   (Revision 2 said "56 ok"; a package tally moves when a package is added, and CLAUDE.md's rule is
   to read the FAIL COUNT. **[r3]**) Re-armed by any `.go` file.
3. **The tune's firing set, pinned by page id, in the default build.** The tuning document is
   uncommittable and `grid300-sel151.txt` holds three scalars per page, which cannot carry geometry.
   The repaired **geometry** is itself committable — rule coordinates are not the document — so
   `testdata/repaired300.txt` records, for each of the 80 pages, the UNFILTERED closed-and-opened
   boxes (D8) in the 300-DPI space, and `TestRepairedFiringSet` runs D8's filters and D3's crossing
   test over that record and demands the firing set match the tune's own LABELS, not a named set of
   page IDs — so a newly-firing page reads as a labeled false positive rather than as a set element
   somebody updates. **[r3.1]**
   Because the record holds pre-filter boxes, §VI.6's deletions turn this red. **What it does not
   pin, stated:** the C half — a change to the closing length moves the boxes and the record is
   stale rather than wrong. `TestRepairedRecordIsCurrent` compares the record's header (which
   carries the `shim.cpp` source hash, the closing length and the per-page DPI — the three things
   the BOXES depend on) against the build, and fails when any moves. **It does NOT carry
   `maxRuleThick` or `minRepairedSpan`: the record is pre-filter and those do not move a box, so
   keying on them would call it stale and demand a C-stack regeneration that changes nothing.
   [r3.1]** Regenerating is a hand step and the command is in the header.
4. **Tagged suite.** `eval "$(./third_party/pins/build-cstack.sh env linux-amd64
~/ocr-runs/644-real/cstack)"`
   then `go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"'
   ./internal/tessocr/ ./internal/fetchcache/` → ok ok.
5. **What the repaired pages ACTUALLY produce, recorded before merge. [r3]** Revision 2 asserted
   p0022 would degrade honestly; its audit showed the claim was unearned and the mechanism made the
   opposite more likely. So for `nbs602-dashed-matrix` and for p0022, the measured `Text`,
   `TextCells`, `TextCellFallback` and `RuleSource` are pasted into §IV D5 before this merges.
   **Accept/reject, decided now rather than once the output is known: a page that publishes a
   `|`-separated table of a FIGURE blocks this change** — it is quotable downstream through
   `cite --ocr-quote` — **until either the lattice test rejects it or a human has read it against
   the page image.** A page that refuses with its reason stated is acceptable and recorded.
   This step also answers §I goal 1 for #933: the levels path either produces boxes on the dashed
   page or is shown not to run there.
6. **Delete-the-row, in the default build.** Remove the thinness filter, remove the span
   requirement, and remove each half of the crossing test — four deletions, each must turn step 3
   or a `lattice_test.go` unit red. All four are Go (D8) and the record is unfiltered (§VI.3), so
   none of them needs the C stack to kill.
7. **Corpus.** `FEOV_OCR_CORPUS=1` + the tagged command on `./internal/fetchcache/ -run
   '^TestCorpusGoldens$'` (`-update` regenerates goldens **and** `STATUS.md`). Expected:
   `nbs602-dashed-matrix` reads `table: true`; **its text may move, and may get worse** — it enters
   the grid branch, where the rotation probe and `TextCells` both act, and its `expect` block's
   `must_contain` list is what measures that. Every *other* page's verdict and text unchanged —
   that is D4's property, and `nbs602-form` and `usfs-birds-markgrid` still detecting is its sharp
   end. Read every changed line; all eight goldens move on the identity hash alone (§V).
8. **Resolution stability, on the gate #1074 built.** The repaired path's Go half is fed normalized
   boxes, so it joins `TestGeometryIsResolutionInvariant`: a synthetic repaired-box fixture
   expressed at 450 and 600 must produce the same verdict and the same table. Pure Go, every CI leg,
   no engine time. **What it pins and what it does not, stated: [r3.1]** it pins the FILTER AND
   CROSSING ARITHMETIC, not "the same dashed page at two resolutions" — the merged test drives
   `TextCells` over synthetic input by design, because tesseract returns different words at
   different resolutions. And a threshold comparison after `round` is invariant only away from the
   cap: a box within ~1 px of `maxRuleThick` flips between scales, so **the fixture is stated as
   off-boundary and a second case sits deliberately ON the cap to record the ±1 px residue** rather
   than a fixture quietly chosen to sit somewhere safe.
9. `go -C scripts run ./check` → read the FAIL COUNT. **`ocr-corpus` does not run there**; step 7
   is run by hand and its result stated in the pull request.

## VII. Risk & Mitigation

Graded likelihood × impact × complexity-to-mitigate. **[r2: the third column is new.]**

- **R1. A figure with dotted guides reads as a table (measured, medium × medium × unknown).**
  p0022. Revision 1 graded the impact low on the strength of a failure direction it had not
  measured; D5 withdraws that. **§VI.5 is now the mitigation and it has a stated accept/reject**: a
  repaired page that PUBLISHES a |-separated table of a figure blocks this change, because
  `cite --ocr-quote` will quote it. Complexity is "unknown" because which way p0022 falls is not yet
  measured, and this plan does not merge until it is. **[r3]**
- **R2. The closing damages recognition (low × high × low).** It must not: the repair is a separate
  binarized copy used ONLY to find rules, never the image handed to `SetImage`. The corpus goldens
  are the check.
- **R3. Cost (low × low × low).** Two extra morphological passes and one connected-component pass
  per page, on the binary image, and only on pages the existing detector rejected (D4). The dump
  crossing the boundary is larger than revision 2's because C filters nothing (D8); that is a few
  hundred boxes of text per page against an engine read measured in seconds.
- **R6. The repair's own constants drift from the normalized space (low × high × low). [r3]** Two
  of the three are 300-DPI numbers converted by hand from a 350-DPI measurement in D1. §VI.1
  re-measures them through the normalized path before implementation, so the conversion is never
  what the tune rests on.
- **R4. The parameters are fitted to one document (medium × medium × medium).** They come from SP
  602's typewriter dashes. The corpus is the guard — the DOE, NBS and USFS pages all constrain them
  — and #1004's breadth is where a second dashed document would come from.
- **R5. `repaired300.txt` goes stale silently (medium × medium × low).** Its boxes are measured
  once and cannot be recomputed in CI. §VI.3's header check — the `shim.cpp` source hash and the
  constants it was generated under — is the loud miss; the residue, stated,
  is that a C-side change is caught by the corpus and not by the 80-page set.

## VIII. What it did, measured

Implemented 2026-09-20. Every number below is from the shipped code, not from the prototype §IV was
written against.

**§VI.5's accept/reject fired, and it blocked the change.** The gate was written before the number
was known: *a page that PUBLISHES a |-separated table of a figure blocks this change.* On the first
run, p0022 did exactly that — `TextCellFallback` empty, `TextCells` non-nil, so the page's text
became a 7-column table with 165 words in 35 cells, built out of a V&V flow diagram's dashed boxes
(`V&V Test Plan`, `V&V Test Design`, `V&V Test Case`…). That is the fabricated table this plan said
it would not ship, and it is quotable through `cite --ocr-quote`.

**D9, the fix, and it coins no constant.** A table's rules BOUND it; a figure's guides sit inside
part of it. Measured, in the 300-DPI space:

| | horizontal stack | tallest vertical | coverage |
|---|---|---|---|
| `nbs602-dashed-matrix`, a real dashed table | y 816–1968 (1152 px) | 1177 px | **102%** |
| IEEE 1012 p0022, a flow diagram | y 675–2481 (1806 px) | 971 px | **54%** |

The separation is not marginal, and the threshold already existed: `coverage = 0.8` in `lattice.go`
is the repository's own constant for this exact idea — "how much of a band a vertical rule must span
to bound its cells". Reusing it keeps one name for one concept and fits nothing to one page.

**The tune, re-measured with the shipped code, at the 80 pages' native 324 DPI:**

| | before | after |
|---|---|---|
| pages the repaired path fires on | p0022 only | **none** |
| the tune's standing false positives | 1 (p0025) | **1 (p0025)** |

**§IV D5 is therefore wrong and is corrected here rather than in place**, so the arithmetic that
produced it stays visible: it predicted the repair would fire on 30 pages, 29 of them already
tables, for a precision of 0.971 → 0.946. The shipped path fires on **zero of the 80**. The
prototype it was measured with used native-pixel constants and no bounding test. **#1027's
acceptance does not need amending after all** — TP=33 FP=1 FN=0 is unchanged — and §I goal 3's
"exactly the pages named in D5 and no others" is met by the empty set.

**The corpus, all eight pages:**

| page | before | after |
|---|---|---|
| `nbs602-dashed-matrix` | `table: false`, read as prose | **`table: true · rules: repair · repaired crossings: 56`**, 1 table, 7 bands, 14 rows, 7 columns, 49 cells, 295 words |
| `nbs602-form`, `usfs-birds-markgrid` | `table: true` | unchanged, `rules: detector` |
| the other five | `table: false` | unchanged, `rules: none` |

D4's property holds exactly: the seven pages this plan does not target changed one line each, the
engine identity hash, and nothing else.

**What the dashed page now reads.** Its row binding is recovered — `| Spot Welding … | 3.7 (2-6) |
4.1 (3-5) | 4.9 (4-6) | 2.1 (1-3) | 2.1 (1-2) | 9.4 (5-6) |` is the matrix's real data, bound to its
row and columns, where before the page read as prose and the binding was lost. Its `expect` block's
`table: want true, got false` is gone.

**What it does NOT read, stated.** The dashed rules are themselves read as text — `ee ee ee`,
`cs cs co ces cee` — and land inside the cells beside the data. That is **#1030**, already open, and
this measurement is its evidence rather than a defect introduced here. The corpus meter still reads
5 of 8 for that reason.

**Delete-the-row: six mutations, six killed.** The thinness cap, the span requirement, each half of
the crossing test, the crossing threshold, and the bounding filter each turn a `repaired_test.go`
case red — all in the default build, no C stack. Two more on the record path: dropping `RuleSource`
from the projection, and never setting it on the receipt, both turn
`TestEveryTablePageNamesItsRuleSource` red.
