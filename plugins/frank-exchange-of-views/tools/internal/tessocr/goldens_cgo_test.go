//go:build tessocr && cgo

package tessocr

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// EVERY LAYOUT CASE, READ BY THE REAL ENGINE, PINNED BYTE FOR BYTE.
//
// The pages here are generated (testdata/gen), so their bytes are reproducible from one file. They
// prove the rules we WROTE; the scans that beat us are the other half of the evidence and live in
// testdata/corpus, committed under the closed licence set in internal/tessocr/corpus. The goldens
// are the WHOLE reading plus what the page measured, so a regression arrives as a diff a human
// reads rather than as a judgement somebody has to make. The layout cases are
// a ruled table of text, rows separated by space inside
// a band, rules the scan broke, a grid of marks, a boxed paragraph, prose with a table in it, two
// tables on one page, and a faint photocopy.
//
// THE GOLDENS PIN A BUILD, NOT A TRUTH. tesseract 5.5.3, leptonica 1.87.0 and the traineddata are
// pinned by the engine identity; bumping a pin moves these files, and reading that diff is the
// review the bump deserves.
//
// Regenerate with:
//
//	go test -tags tessocr -run TestTextCellGoldens -update ./internal/tessocr/
var update = flag.Bool("update", false, "rewrite the layout-case goldens from what the engine reads now")

func TestTextCellGoldens(t *testing.T) {
	pngs, err := filepath.Glob("testdata/*.png")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(pngs)
	cases := 0
	for _, p := range pngs {
		if filepath.Base(p) == "gridcrop.png" { // the detector's own fixture, not a layout case
			continue
		}
		cases++
		t.Run(strings.TrimSuffix(filepath.Base(p), ".png"), func(t *testing.T) {
			png, rerr := os.ReadFile(p)
			if rerr != nil {
				t.Fatal(rerr)
			}
			en, nerr := New()
			if nerr != nil {
				t.Fatal(nerr)
			}
			defer en.Close()
			res, perr := en.ReadPage(png, Grid300)
			if perr != nil {
				t.Fatal(perr)
			}
			got := renderGolden(res)
			golden := strings.TrimSuffix(p, ".png") + ".golden"
			if *update {
				if werr := os.WriteFile(golden, []byte(got), 0o644); werr != nil {
					t.Fatal(werr)
				}
				t.Logf("wrote %s (%d bytes)", golden, len(got))
				return
			}
			want, gerr := os.ReadFile(golden)
			if gerr != nil {
				t.Fatalf("no golden for this page (%v) — run with -update and READ what it wrote", gerr)
			}
			if got != string(want) {
				t.Errorf("the reading of %s moved.\n\n--- golden\n%s\n--- now\n%s", p, want, got)
			}
		})
	}
	if cases == 0 {
		t.Fatal("no layout-case pages at all — an empty set matches forever and pins nothing")
	}
}

// renderGolden is the page's whole reading, with the facts the record would carry above it: the
// path taken and what it measured. A golden that held only the text would go green on a page that
// silently stopped reconstructing.
func renderGolden(res PageResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "table: %v · rotated: %v · grid h=%d v=%d intersections=%d\n",
		res.Table, res.RotatedPage, res.Grid.HPix, res.Grid.VPix, res.Grid.Intersections)
	switch {
	case res.Reconstruction != nil:
		st := res.Reconstruction
		fmt.Fprintf(&b, "path: marks (%d placed of %d, %d columns, %d subcolumns, %d rows, levels read %d, unread %d)\n",
			st.MarksPlaced, st.MarksTotal, st.ColumnsFound, st.SubColumnsFound, st.RowsFound, st.LevelsRead, st.LevelsUnread)
	case res.TextCells != nil:
		st := *res.TextCells
		fmt.Fprintf(&b, "path: text cells (%d columns, %d ruled bands, %d rows, %d cells, %d words placed, %d from the sparse pass, %d around the table, %d loose)\n",
			st.Columns, st.Bands, st.Rows, st.Cells, st.WordsPlaced, st.WordsFromSparse, st.WordsAround, st.WordsLoose)
	default:
		fmt.Fprintf(&b, "path: plain text\n")
	}
	if res.Fallback != "" {
		fmt.Fprintf(&b, "marks fell back: %s\n", res.Fallback)
	}
	if res.TextCellFallback != "" {
		fmt.Fprintf(&b, "cells fell back: %s\n", res.TextCellFallback)
	}
	if res.Evidence.RefusedTable != "" {
		b.WriteString("--- the reconstruction the dropout gate refused (evidence, not the reading)\n")
		b.WriteString(res.Evidence.RefusedTable)
	}
	b.WriteString("---\n")
	b.WriteString(res.Text)
	if !strings.HasSuffix(res.Text, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}
