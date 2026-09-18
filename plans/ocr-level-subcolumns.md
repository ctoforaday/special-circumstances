# Table 2's level subcolumns

Status: gblock ruled the scope 2026-09-17 — subcolumns only, proven on generated goldens, the
dropout gate unchanged. Carries #933.

## I. Summary & Goals

**The defect.** Mark reconstruction groups raw mark columns under repeated "Levels" captions into
supercolumns and emits `X X X` per activity. On IEEE 1012's Table 2 every activity has four
integrity-level subcolumns (4, 3, 2, 1), and the level is the table's content: which levels require
a task. The merged cell cannot say it even when every mark is found.

**Measured basis (2026-09-17, real engine, p51 rotated).**

- The ten "Levels" captions read in the page TSV (y 735). The `4 3 2 1` row under them does not:
  one digit of forty. A re-read of the band under the captions reads none, in three segmentation
  modes.
- The digits are bold and boxed tightly by vertical rules. Read one CELL at a time — the lattice
  from #932 gives the boxes — **32 of 32 cells read**. Where the lattice merged two subcolumns the
  read returns both digits with a separator (`4|3`, `2/1`), which says how many levels the cell
  spans.

**Goals.**

1. A mark table whose supercolumns carry level subcolumns emits one column per (activity, level),
   headed `<activity> L<level>`, and a cell is `X` or empty.
2. The level of a subcolumn is READ, never inferred from position or mark count. A subcolumn whose
   level was not read is headed `L?` and counted.
3. Tables without level subcolumns emit exactly what they emit today (p54 byte-identical).
4. The dropout gate is untouched, so no real Table 2 page's text changes until dropout is fixed —
   but the reconstruction's stats, recorded even on a fallback, show the levels recovered.

**Non-goals.** Making Table 2 pass the gate. Row dropout on Table 2. Tables whose subcolumns carry
no printed labels (they keep supercolumn grouping).

## II. Technical Context

`Reconstruct(tsv, headers)` is pure: mark columns come from clustering mark x-centres, supercolumns
from `findAnchors` (a caption repeated ≥4 times at a supercolumn pitch), and `bandHeaders` names
them. `readGridPage` holds the pixels, the engine and (since #932) `GridLines`.

## III. Proposed Changes

**D1. Levels are read from the header cells the lattice boxes.** In the grid branch, when the TSV
carries a repeated caption (the same test `findAnchors` applies), the band of lattice cells directly
below that caption line and above the grid is the level band. Each cell is cropped inside its rules
with a white margin and read alone as a single character run. *Beaten:* the page TSV (1 of 40
digits), a band re-read (0 of 40), inferring 4-3-2-1 by position (the issue forbids it: a table need
not print its levels in that order or contiguously).

**D2. A cell that read several digits spans that many subcolumns.** Its width is divided equally,
left to right, one digit per part — the lattice missed a rule there, and the digits say how many it
missed. A cell that read something other than digits labels nothing.

**D3. The COLUMNS come from the level boxes, and marks are placed into them.** Reconstruct takes
the boxes as an input (`[]LevelBox{X0, X1, Label}`) beside `headers`, so it stays pure and testable
without the engine. Mark columns cannot supply the output columns: a level no row marks leaves no
mark to cluster, and on Table 2 the marks reveal 9, 31, 14 and 3 of the ~40 printed subcolumns
(p50-53). The printed header can. So with anchors and level boxes, each box is an output column
under the activity whose anchor is nearest, and each mark lands in the box containing its centre.

*Correction to the scoping message of 2026-09-17:* the dropout gate already counts the raw
subcolumns the marks reveal (`ExpectedIntersections` uses `SubColumnsFound`), not the ten merged
activity columns; Table 2's excess crossings are dropout alone. The ruling — gate unchanged — stands.

**D5. A read stands only where the table agrees, with a quorum.** Every activity prints the same
level header, so the reads at one position are independent witnesses of one printed digit. A read
that disagrees with the majority at its position is made UNREAD — never replaced by the majority,
which would be a level nobody read in that cell — and the majority needs at least half the
activities to have read that position. Measured without it: p52 read one "1" as "4", and on p53 a
lone misaligned box was its own majority. Also forced by the page: a band cell is a level cell only
under the captions (the row-label column's header read as garbage), and a cell the lattice merged
is split by the band's cell width BEFORE it is read (read across the missing rule, it returned
"3|2|4" for 3 2 1).

**Measured, final (real engine, IEEE 1012, levels checked against the printed 4 3 2 1):**

| page | levels read | unread | wrong |
|---|---|---|---|
| p50 | 40 | 0 | 0 |
| p51 | 40 | 0 | 0 |
| p52 | 36 | 4 | 0 |
| p53 | 13 | 20 | 0 |

p51's refused table, checked row by row against the image for its first two rows: every mark sits
under the right activity and level.

**D6. The refused reconstruction is kept as evidence.** The gate still refuses Table 2, so its
per-level table would otherwise exist nowhere. It is written beside the reading as
`p%04d.refused.md`, like the TSVs — debugging evidence, not the reading, in no hash.

**D4. The fields.** `Stats` gains `LevelsRead` (subcolumns with a read level) and `LevelsUnread`
(subcolumns headed `L?`).

**Components.**

- `[MODIFY] internal/tessocr/reconstruct.go`: `LevelBox`, the per-subcolumn emit, the stats.
- `[NEW] internal/tessocr/levels.go`: the level band from the lattice and the TSV, D2's split.
- `[MODIFY] internal/tessocr/page.go`: read the band's cells, pass the boxes in.
- `[MODIFY] internal/tessocr/testdata/gen/pages.go`: a Table-2-shaped page (activities with
  `Levels` captions and a `4 3 2 1` row, bold marks), and its golden.
- Tests: pure-Go reconstruct tests with level boxes; the golden through the real engine.

## IV. Risk & Mitigation

- **R1. A digit misread as another digit (low × high).** The header is `L3` where the scan says 4.
  Single-cell reads of bold digits measured 32/32; the golden pins the generated page, and the
  corpus harness (#1004) is where a second document tests it.
- **R2. The level band is found on a table that is not Table 2 (low × medium).** It needs the
  repeated-caption test to fire first, which is today's supercolumn condition; a table without it is
  unchanged.
- **R3. Cost (low × low).** One tiny crop read per header cell, only on pages with anchors: 40 on a
  Table 2 page.

## V. Verification Plan

1. `go test -count=1 ./...` and the tagged engine suite.
2. Pure-Go: a reconstruction with level boxes emits one column per level; a merged cell splits; an
   unread level is `L?` and counted; no anchors → byte-identical to today.
3. Golden: the generated Table-2-shaped page, whole reading pinned.
4. Real document: p54 byte-identical; p50–53's recorded stats show `levels_read` near 40 and the
   text unchanged (the gate still falls them back).
5. Delete-the-row on D2's split and the `L?` path.
6. `go -C scripts run ./check`.
