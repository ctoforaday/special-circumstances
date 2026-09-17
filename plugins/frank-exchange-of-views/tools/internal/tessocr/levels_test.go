package tessocr

import (
	"fmt"
	"strings"
	"testing"
)

// THE LEVEL HEADER (#933), pinned without the C stack: header cells and their reads go in, level
// boxes come out. The real engine's reads are measured in the plan (IEEE 1012 p50-53: 40, 40, 36
// and 13 levels read, none wrong).

func labels(bs []LevelBox) string {
	var out []string
	for _, b := range bs {
		if b.Label == "" {
			out = append(out, "?")
		} else {
			out = append(out, b.Label)
		}
	}
	return strings.Join(out, "")
}

func TestLevelBoxesSplitACellThatReadSeveralDigits(t *testing.T) {
	cells := []Cell{{X0: 0, X1: 100}, {X0: 100, X1: 150}, {X0: 150, X1: 200}}
	got := LevelBoxes(cells, []string{"4|3", "2", "no digit here"})
	if labels(got) != "432?" {
		t.Fatalf("labels = %q, want 432? — a two-digit cell is two levels, a cell with no digit is one unread level", labels(got))
	}
	if got[0].X1 != 50 || got[1].X0 != 50 {
		t.Errorf("the two-digit cell split at %d/%d, want 50 — its width divides equally", got[0].X1, got[1].X0)
	}
}

func TestSplitWideCellsDividesAMergedCellByTheBandsCellWidth(t *testing.T) {
	band := []Cell{{X0: 0, X1: 50}, {X0: 50, X1: 100}, {X0: 100, X1: 250}, {X0: 250, X1: 300}}
	got := splitWideCells(band)
	if len(got) != 6 {
		t.Fatalf("cells = %d, want 6 — a 150 px cell in a band of 50 px cells is three", len(got))
	}
	if got[2].X0 != 100 || got[2].X1 != 150 || got[4].X1 != 250 {
		t.Errorf("split at %d-%d … %d, want 100-150 … 250", got[2].X0, got[2].X1, got[4].X1)
	}
}

// ten activities, each 4 3 2 1, one misread and one isolated fifth box
func tableOfLevels(misread map[int]string, extra bool) ([]LevelBox, []float64) {
	var boxes []LevelBox
	var centres []float64
	want := []string{"4", "3", "2", "1"}
	for a := 0; a < 10; a++ {
		base := 1000 * a
		centres = append(centres, float64(base+100))
		for k := 0; k < 4; k++ {
			lab := want[k]
			if m, ok := misread[a*4+k]; ok {
				lab = m
			}
			boxes = append(boxes, LevelBox{X0: base + 50*k, X1: base + 50*k + 50, Label: lab})
		}
	}
	if extra {
		// a fifth box grouped under activity 6 — the next activity's "4", misaligned by the lattice
		boxes = append(boxes, LevelBox{X0: 6000 + 200, X1: 6000 + 250, Label: "4"})
	}
	return boxes, centres
}

func TestAgreeWithTableUnreadsAMisreadNeverReplacingIt(t *testing.T) {
	boxes, centres := tableOfLevels(map[int]string{3: "4"}, false) // activity 0 reads its "1" as "4"
	got := AgreeWithTable(boxes, centres)
	if got[3].Label != "" {
		t.Errorf("the misread 4 at a position the table reads as 1 stood as %q — it must be unread, never corrected to 1", got[3].Label)
	}
	for i, b := range got {
		if i != 3 && b.Label == "" {
			t.Errorf("box %d lost a read the table agrees with", i)
		}
	}
}

func TestAgreeWithTableNeedsAQuorum(t *testing.T) {
	boxes, centres := tableOfLevels(nil, true)
	got := AgreeWithTable(boxes, centres)
	if last := got[len(got)-1]; last.Label != "" {
		t.Errorf("a lone read at a position no other activity has stood as %q — one witness is not a majority of ten", last.Label)
	}
}

func TestReconstructEmitsOneColumnPerLevel(t *testing.T) {
	// Four activities with a "Levels" caption and a name over each, two levels apiece, and two rows
	// of marks. No rotated-band headers are passed: those name the real columns and would merge the
	// level subcolumns away, and a table with printed levels takes its names from its own band.
	var ws []word
	for a, name := range []string{"Acq", "Sup", "Dev", "Ops"} {
		ws = append(ws, word{name, 330 + 400*a, 40, 60, 24, 95})
	}
	caption := func(x int) word { return word{"Levels", x, 100, 80, 24, 95} }
	for a := 0; a < 4; a++ {
		ws = append(ws, caption(300+400*a))
	}
	ws = append(ws, word{"Planning", 20, 400, 110, 24, 95}, word{"Review", 20, 500, 90, 24, 95})
	for a := 0; a < 4; a++ {
		ws = append(ws, word{"X", 300 + 400*a, 400, 20, 24, 90}) // level A of each activity, row 1
		ws = append(ws, word{"X", 500 + 400*a, 500, 20, 24, 90}) // level B, row 2
	}
	var levels []LevelBox
	for a := 0; a < 4; a++ {
		levels = append(levels, LevelBox{X0: 260 + 400*a, X1: 400 + 400*a, Label: "2"},
			LevelBox{X0: 400 + 400*a, X1: 600 + 400*a, Label: ""})
	}
	table, st, err := Reconstruct(fakeTSV(ws), nil, levels)
	if err != nil {
		t.Fatal(err)
	}
	head := strings.SplitN(table, "\n", 2)[0]
	for _, want := range []string{"Acq L2", "Acq L?", "Ops L2", "Ops L?"} {
		if !strings.Contains(head, want) {
			t.Errorf("header %q does not name %q", head, want)
		}
	}
	if st.LevelsRead != 4 || st.LevelsUnread != 4 || st.ColumnsFound != 8 {
		t.Errorf("stats read=%d unread=%d columns=%d, want 4, 4, 8", st.LevelsRead, st.LevelsUnread, st.ColumnsFound)
	}
	if row := rowOf(table, "Planning"); strings.Count(row, "X") != 4 || !strings.HasPrefix(strings.TrimSpace(strings.Split(row, "|")[2]), "X") {
		t.Errorf("row Planning = %q, want an X under each activity's first level", row)
	}
}

func rowOf(table, label string) string {
	for _, l := range strings.Split(table, "\n") {
		if strings.HasPrefix(l, "| "+label+" ") {
			return l
		}
	}
	return ""
}

// THE LEVEL BAND IS THE CELLS UNDER THE CAPTIONS. The same band crosses the row-label column,
// whose header ("Software Integrity Levels") is not a level; read, it becomes a bogus unread
// column at the table's left edge, which is what p51 emitted before this rule.
func TestLevelBandCellsAreTheCellsUnderTheCaptions(t *testing.T) {
	var rules []string
	// a header band (y 100-200) holding the captions, the level band (200-260), one grid row
	for _, y := range []int{100, 200, 260, 400} {
		rules = append(rules, fmt.Sprintf("h 0 %d 1700 3", y))
	}
	// the label column 0-500, then four activities of four levels, 50 px each, from x 500
	rules = append(rules, "v 0 100 3 303", "v 500 100 3 303")
	for a := 0; a < 4; a++ {
		base := 500 + 300*a
		for k := 1; k <= 4; k++ {
			rules = append(rules, fmt.Sprintf("v %d 200 3 203", base+50*k))
		}
		rules = append(rules, fmt.Sprintf("v %d 100 3 303", base+300))
	}
	var ws []word
	for a := 0; a < 4; a++ {
		ws = append(ws, word{"Levels", 560 + 300*a, 150, 80, 24, 95})
	}
	ws = append(ws, word{"Software", 40, 220, 90, 24, 95})
	cells := LevelBandCells(ParseLattice(ruleDump(rules...)), fakeTSV(ws))
	if len(cells) == 0 {
		t.Fatal("no level band found under four repeated captions")
	}
	for _, c := range cells {
		if c.X1 <= 500 {
			t.Errorf("the row-label column's cell (%d-%d) was taken as a level cell", c.X0, c.X1)
		}
		if c.Y0 != 201 {
			t.Errorf("a level cell sits at y %d, want the band directly under the captions (201)", c.Y0)
		}
	}
}
