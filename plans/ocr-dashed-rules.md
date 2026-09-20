# A rule drawn as dashes

Status: revision 2, measured 2026-09-20. **BLOCKED on #1074, and revision 3 is owed.** Carries #1027.
Revision 1 failed its audit on eight gaps; what each changed is marked **[r2]** where the
design moved, so a reader can see which parts were argued rather than written.

## 0. What the second audit found, and why this stopped

Revision 2 also failed, on a finding that invalidates §IV D6 — the fix revision 1's audit
demanded. **D6 repairs a gate this plan's pages never reach.**

- `nbs602-dashed-matrix` is a numeric matrix. `isStrongMark` (`reconstruct.go:203`) needs an
  x/X token, so `Reconstruct` returns `ErrNoMarks`, and `fallbackReason` (`page.go:329-352`)
  **returns at its first case** — the `MaxIntersectionRatio` branch at line 344 is never
  evaluated. Filling `GridStats` with a repaired measurement changes nothing on a dashed page.
- The gate that actually decides whether a repaired page's text is published is `TextCells`
  (`page.go:273`), refused by `CellStats.refuse()` (`lattice.go:702`) on
  `MinTextCellColumns` / `MinTextCellRows` / `MinWordsPerCell`.
- Those thresholds, and the `Lattice.Tables()`/`Cells()` geometry they run on — `ruleJoin`,
  `minBandHeight`, `minRowGap`, `coverage` — are **pixels at 300 DPI, applied unscaled**. The
  dashed page renders at 350 and the tune measurement ran at 324. That is #1074, filed
  2026-09-20 with the measurement behind it: one physical table read at 300/350/400/501/600
  reports 5 different shapes, and scaling those constants by dpi/300 restores the 501 read to
  the 300-DPI structure exactly.
- `MaxIntersectionRatio = 4.0` is in the same class: its numerator scales with DPI² while
  `ExpectedIntersections()` is DPI-invariant, so 4.0 is a one-resolution fit stated as a
  constant across a 300-600 band.

**Therefore #1074 comes first.** Fitting three new constants (§IV D1) on top of a geometry
layer that is wrong at the dashed page's own resolution would fit them to the error. Revision 3
is written after #1074 lands, and re-aims D6 and §VI.4 at `refuse()`.

Also owed in revision 3, from the same audit: §V's census is missing eight consumers its own
sweep returns (`lattice_test.go`, `levels.go`/`levels_test.go`, `pageread_test.go`,
`engine_cgo_test.go`, `textcells_cgo_test.go`, `fetchocr_test.go`, `testdata/gen/pages.go`);
§VI.2's record must hold the C entry point's **unfiltered** components or the delete-the-row
on D8's filters cannot turn it red; D7's failure contract covers the geometry and not the
three counts, and `ParseLattice` skips a line it cannot parse, which revives the silent zero;
and D6's source field needs a statement that its zero value denotes no path.

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
   consulted only where the existing one found nothing (§III D4). **[r2]**
3. On the tuning document, the repaired path adds **exactly the pages named in §III D5 and no
   others**. It does add one. Goal 3 in revision 1 said "the verdicts do not move", which
   contradicted D5 and contradicts nothing now: the criterion is *the named exception, and no
   unnamed one*. #1027's acceptance ("TP=33 FP=1 FN=0 … do not regress") is amended in the issue
   before this ships, or this plan does not ship. **[r2]**
4. A page detected by the repaired path carries a grid measurement that is **its own**, in the units
   the acceptance gate is calibrated in — never the existing path's zero (§III D6). **[r2]**

**Non-goal.** Reconstructing SP 602's dashed tables well. Detection is this plan; what the
reconstruction then does with rules it can finally see is measured after, on the corpus.

**The value case, weighed. [r2]** The machinery is real: a new C entry point, three new constants in
the per-page derivation, a second lattice source, a second measurement feeding the acceptance gate.
Against it: a 60-page document reads twelve of its tables as prose today, and the class is not
one document — typewriter and teletype rules are dashes by construction, so every scanned technical
report of that era is in it. The naive alternative (ship nothing, accept the residue) was the
status quo through #934 and is what #1027 exists to overturn. The fork is recorded because it is a
real one, not because it is close.

## II. Technical Context **[r2]**

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

**D1. Repair, then measure what survived.** Three steps on the binarized page, per axis:

1. **close** by the dash gap — joins a rule's dashes, and unavoidably the characters of a word;
2. **open** by the rule length (the existing `ruleRunInches`, unchanged and not re-coined) — keeps
   long runs, of which a closed line of prose is still one;
3. **keep only components thin in the perpendicular axis** — a rule is thin; a line of prose closed
   into a run is as tall as its x-height.

The three new lengths, stated as this file states lengths — **measured pixels over the DPI they were
measured at**, with the kind named **[r2]**:

| constant | kind | measured | derived |
|---|---|---|---|
| `dashGapInches` | a length ∝ DPI | 25 px at 350 | `25.0 / 350` |
| `maxRuleThickInches` | a length ∝ DPI | 12 px at 350 | `12.0 / 350` |
| `minRepairedSpanInches` | a length ∝ DPI | 600 px at 350 | `600.0 / 350` |

Measured on the dashed matrix: this recovers **8 horizontal rules** at y = 953, 1195, 1386, 1579,
1722, 1914, 2105, 2296 — gaps of 242, 191, 193, 143, 192, 191, 191, the 143 being the header band,
so they are regularly and not evenly spaced **[r2]** — and the vertical rules that bound its
columns, where the current detector recovers nothing at all.

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
downstream through `cite --ocr-quote`. **Before this merges, p0022's actual `Text`, `Fallback` and
`TextCellFallback` are run and recorded in this section.** If it fabricates a table, the design
needs a guard this plan does not yet contain, and the guard is designed then — not asserted now.

**D6. A repaired page measures its own grid, in the gate's units. [r2]** This is the gap revision 1
would have shipped. `fallbackReason` (`page.go:344`) discards a reconstruction whose grid measures
more than `MaxIntersectionRatio` times the lattice the reconstruction accounts for — the safety net
against total row dropout. It reads `grid.Intersections`, which on a repaired page is **0**, so
`0 > 4.0 × expected` is false forever: on exactly the new class of pages the net would be dead, and
a shredded reconstruction would be published as the page text with nothing said.

The fix is not a second gate with a second constant. `MaxIntersectionRatio` is calibrated on
crossing **pixels** (the AND of two morphological openings), not on lattice points, and the repaired
masks are two openings whose AND is the same kind of measurement. So `tessocr_repaired_rules`
returns the repaired masks' `hpix`, `vpix` and `inter` alongside the geometry, those fill the
`GridStats` the page carries, and the existing gate stays live with its existing constant. Two
consequences, both discharged in §VI:

- whether **4.0** is the right boundary in repaired units is a measurement, not an inheritance —
  §VI.4 measures the ratio on every page the repaired path fires on;
- `GridStats` gains a field naming **which path measured it**, because "this page's rules were
  recovered by repair" is a fact the record's readers act on, and inferring it from a zero is
  exactly the unmediated read `facts-are-fields` refuses.

**D7. Failure contract. [r2]** `tessocr_repaired_rules` follows its siblings: `nullptr` is a
failure (decode, binarize, morphology or allocation), an empty geometry string is a *measurement* —
a page with no repaired rules. Go turns the first into an error that reaches the receipt and the
second into an empty lattice. Because D4 makes the repair a second opinion, a repair *failure* would
otherwise degrade to "not a table", which is indistinguishable from "no dashed rules here" on the
one class this plan exists for.

**D8. The thin/span filters and the crossing test live in Go. [r2]** C returns the closed-and-opened
components as boxes; the thinness cap, the span requirement and D3's crossing test are Go, so they
are unit-testable in the default build and each is a deletable row (§VI.6) that does not need the
C stack to kill.

## V. Components

**Census — every carrier of a grid fact, swept with**
`grep -rn "res\.Table\|\.Grid\b\|GridStats\|GridIntersections\|GridLines\|Lattice\|TextCells(" plugins/frank-exchange-of-views/tools --include=*.go` **[r2: revision 1 had no census]**

- `[MODIFY] internal/tessocr/shim.{h,cpp}` — `tessocr_repaired_rules(png, len, close, sel)` returning
  the surviving components' boxes plus the repaired masks' hpix/vpix/inter, with D7's contract.
- `[MODIFY] internal/tessocr/engine_cgo.go` **and** `internal/tessocr/engine_stub.go` — both halves,
  or the untagged build does not compile.
- `[MODIFY] internal/tessocr/tessocr.go` — D1's three lengths join `GridFor(dpi)`; `GridStats` gains
  D6's source field.
- `[MODIFY] internal/tessocr/lattice.go` — D3's crossing test and D8's filters, over repaired boxes.
- `[MODIFY] internal/tessocr/page.go` — D4's second-opinion ordering; `fallbackReason` unchanged,
  and §VI.4 is what proves it still fires.
- `[MODIFY] internal/fetchcache/pageread.go` — `pageReceipt.GridIntersections` /
  `PageReading.GridIntersections` (lines 145-151, 187, 200, 288-294) carry the repaired measurement
  and the source; their doc comment states the dropout gate and must say which path measured.
- `[MODIFY] internal/tessocr/goldens_cgo_test.go:80` and `internal/fetchcache/corpus_cgo_test.go:174`
  — `renderGolden` prints `grid h=%d v=%d intersections=%d`; the source belongs on that line, or
  every golden reads as if one path produced it.
- `[MODIFY] internal/tessocr/detector_test.go` — p0022 joins the boundary set with its **measured**
  outcome (D5), not an asserted class.
- `[MODIFY] internal/tessocr/testdata/repaired350.txt` (NEW) — §VI.2's record.
- `[MODIFY] testdata/corpus/*` — `nbs602-dashed-matrix` changes verdict; its `expect` already
  demands `table: true`. All eight `reading.golden` files and `STATUS.md` regenerate regardless,
  because line 1 carries `DefaultPageEngine.Identity()` and that hashes this package's source.
- `[MODIFY] internal/cli/ocr.go:94` — `ocr pages --help` still says "The default 300 DPI is the OCR
  engine's operative resolution". Stale since #1031, and wrong again here. **[r2]**

## VI. Verification Plan

Each step in its own shell; the cstack environment leaking into a golden step makes it fail.

1. **Default build.** `go build ./... && go vet ./... && go test -count=1 ./...` → 56 ok.
   Re-armed by any `.go` file.
2. **The tune's firing set, pinned by page id, in the default build. [r2]** Revision 1 promised a
   pin that could not exist: the pages are uncommittable and `grid300-sel151.txt` holds three
   scalars, which cannot carry geometry; and it pinned the *cardinality* (30), which stays green
   when one page leaves the set and another joins. Instead, the per-page repaired **geometry** is
   itself committable — rule coordinates are not the document — so `testdata/repaired350.txt`
   records, for each of the 80 pages, the surviving boxes and the repaired hpix/vpix/inter, and
   `TestRepairedFiringSet` re-runs D3's crossing test over that record and demands the firing set
   equal the named set of page ids. This pins the Go half (D8) by construction and by id, not by
   count. **What it does not pin, stated:** the C half — a change to the closing length moves the
   boxes, and the record is then stale rather than wrong. `TestRepairedGeometryIsCurrent` compares
   the record's declared constants against `GridFor(324)` and fails when they drift; the boxes
   themselves are re-measured by hand and the regeneration command is in the record's header.
   Re-armed by: `shim.cpp`, the three constants, `lattice.go`.
3. **Tagged suite.** `eval "$(./third_party/pins/build-cstack.sh env linux-amd64 $CSTACK)"` then
   `go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"'
   ./internal/tessocr/ ./internal/fetchcache/` → ok ok. **[r2: revision 1 said "and the tagged
   suite" and named no command.]**
4. **The dropout gate still fires, in repaired units (D6).** For every page the repaired path
   fires on, record `grid.Intersections / st.ExpectedIntersections()`. The ratio must sit below 4.0
   on pages that reconstruct correctly and above it on a page with induced row dropout; if the
   repaired distribution does not straddle 4.0, the constant is re-measured **and this plan is
   revised before it ships** rather than the gate being left nominally alive. Re-armed by
   `shim.cpp`, `page.go`.
5. **Corpus.** `FEOV_OCR_CORPUS=1` + the tagged command on `./internal/fetchcache/ -run
   TestCorpusGoldens` (`-update` regenerates goldens **and** `STATUS.md` together). Expected:
   `nbs602-dashed-matrix` reads `table: true`; **its text may move, and may get worse** — it enters
   the grid branch, where the rotation probe and `TextCells` both act, and its `expect` block's
   `must_contain` list is what measures that **[r2: revision 1's "no page's text moves" was false by
   design]**. Every *other* page's verdict and text unchanged — that is D4's property, and
   `nbs602-form` and `usfs-birds-markgrid` still detecting is its sharp end.
6. **Delete-the-row.** Remove the thinness filter, the span requirement, and each half of the
   crossing test in turn; each must fail the suite. All four are Go (D8), so each kills in the
   default build.
7. **Resolution stability.** As #1031 measured the old constants across 300/400/600, measure the
   repaired verdict on the corpus at the ends of that range. A verdict that depends on the render
   DPI is a constant stated in the wrong units.
8. `go -C scripts run ./check`.

## VII. Risk & Mitigation

Graded likelihood × impact × complexity-to-mitigate. **[r2: the third column is new.]**

- **R1. A figure with dotted guides reads as a table (measured, medium × medium × unknown).**
  p0022. Revision 1 graded the impact low on the strength of a failure direction it had not
  measured; §IV D5 now withdraws that, and §VI.4 plus D5's own measurement are what set this row's
  impact. Until they run, the mitigation is not designed.
- **R2. The closing damages recognition (low × high × low).** It must not: the repair is a separate
  binarized copy used ONLY to find rules, never the image handed to `SetImage`. The corpus goldens
  are the check.
- **R3. Cost (low × low × low).** Two extra morphological passes and one connected-component pass
  per page, on the binary image, and only on pages the existing detector rejected (D4).
- **R4. The parameters are fitted to one document (medium × medium × medium).** They come from SP
  602's typewriter dashes. The corpus is the guard — the DOE, NBS and USFS pages all constrain them
  — and #1004's breadth is where a second dashed document would come from.
- **R5. `repaired350.txt` goes stale silently (medium × medium × low). [r2]** Its boxes are measured
  once and cannot be recomputed in CI. §VI.2's constants check is the loud miss; the residue, stated,
  is that a C-side change is caught by the corpus and not by the 80-page set.
