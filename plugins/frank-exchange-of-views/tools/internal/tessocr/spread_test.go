package tessocr

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The fixtures are the engine's own TSV of the corpus page nbs602-dashed-matrix (SP 602 p44, the
// page #1030 was filed from) at its native 350.21 DPI, under both segmentation modes, and
// tesseract's own text output of the same recognition — read by this build of the C stack before
// the plain reading was written from the TSV.
func dashedMatrix(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "dashedmatrix."+name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// dataValue is a cell of the matrix as printed: a mean and its range, "3.7" and "(2-6)".
var dataValue = regexp.MustCompile(`^\(?\d\.\d\)?$|^\(\d-\d\)$`)

// THE RULES GO AND THE VALUES STAY. Every line either pass loses must hold no value of the matrix
// and no word of its labels, every value either pass read must survive, and no line left behind may
// still be a dashed rule read as "ee".
func TestTheDashedMatrixLosesItsRulesAndKeepsItsValues(t *testing.T) {
	for _, tc := range []struct {
		pass        string
		wantRemoved int
	}{
		// Counted by hand from the fixtures: the auto pass reads three rules as lines of their own
		// ("9S SS SSS SS 9539 be re rer f rrr fl rr ff = ="); the sparse pass reads nineteen.
		{"auto.tsv", 3},
		{"sparse.tsv", 19},
	} {
		t.Run(tc.pass, func(t *testing.T) {
			read := parseTSVWords(dashedMatrix(t, tc.pass))
			kept, removed := removeSpreadLines(read)
			if len(removed) != tc.wantRemoved {
				t.Errorf("removed %d lines, want %d:\n%s", len(removed), tc.wantRemoved, strings.Join(removed, "\n"))
			}
			labels := []string{"Spot", "Welding", "Arc", "Drilling", "Small", "Assembly", "Machine", "Press", "Loading", "Touch", "Simple", "Complex", "Vision", "Force"}
			for _, ln := range removed {
				for _, f := range strings.Fields(ln) {
					if dataValue.MatchString(f) {
						t.Errorf("a removed line held the value %q: %q", f, ln)
					}
				}
				for _, l := range labels {
					if strings.Contains(ln, l) {
						t.Errorf("a removed line held the label %q: %q", l, ln)
					}
				}
			}
			values := func(ws []tsvWord) int {
				n := 0
				for _, w := range ws {
					if dataValue.MatchString(w.text) {
						n++
					}
				}
				return n
			}
			if got, want := values(kept), values(read); got != want || want == 0 {
				t.Errorf("%d of the %d values read survived", got, want)
			}
			for _, ln := range layoutLines(kept) {
				ee := 0
				for _, w := range ln {
					if strings.Trim(w.text, "!|") == "ee" || w.text == "eee" {
						ee++
					}
				}
				if ee >= 3 {
					t.Errorf("a dashed rule is still in the reading: %q", lineText(ln))
				}
			}
		})
	}
}

// THE PLAIN READING IS TESSERACT'S OWN TEXT, written from the TSV. Measured on all 13 corpus pages
// before tesseract's text output stopped being called; pinned here on the one page whose TSV is a
// fixture, so a change to how proseText joins lines and paragraphs cannot pass unseen.
func TestProseTextIsTesseractsOwnTextForTheSameRecognition(t *testing.T) {
	got := NormalizeReading(proseText(parseTSVWords(dashedMatrix(t, "auto.tsv"))))
	want := NormalizeReading(dashedMatrix(t, "auto.txt"))
	if got != want {
		t.Errorf("proseText differs from tesseract's text output.\n--- tesseract\n%s\n--- proseText\n%s", want, got)
	}
}

// THE NO-GRID PATH LOSES ITS SPREAD LINES AND SAYS SO. proseResult is everything that path does after
// recognition; driven with the dashed matrix's own TSV it must remove the three rules the auto pass
// read, state them, and keep the print.
func TestAPageWithNoGridLosesItsRuleLinesOnTheRecord(t *testing.T) {
	res := proseResult(dashedMatrix(t, "auto.tsv"), GridStats{})
	if len(res.RemovedLines) != 3 || !strings.HasPrefix(res.RemovedLines[1], "9S SS SSS SS 9539") {
		t.Errorf("removed %q, want the three rule lines", res.RemovedLines)
	}
	for _, ln := range res.RemovedLines {
		if strings.Contains(res.Text, ln) {
			t.Errorf("a removed line is still in the reading: %q", ln)
		}
	}
	for _, want := range []string{"Spot Welding", "Machine Tool i 2.8", "Complex vision - recognize shapes"} {
		if !strings.Contains(res.Text, want) {
			t.Errorf("the reading lost %q:\n%s", want, res.Text)
		}
	}
}

// THE TEXT-CELL FLOOR IS JUDGED ON WHAT THE PASSES READ, NOT ON WHAT WAS KEPT. usgs-landscape-table's
// cells hold 643 words as read and 242 once its leaders and rules are gone — 4.23 and 1.59 per cell
// across 152 cells. Judged on the kept count it reads as plain text and loses values its cells
// hold ("40,173", "2,890", "3,247"); the corpus record now pins those.
func TestTheTextCellFloorIsJudgedOnTheWordsRead(t *testing.T) {
	landscape := CellStats{Tables: 1, Bands: 10, Rows: 10, Columns: 16, Cells: 152, WordsPlaced: 242, WordsRead: 643}
	if why := landscape.refuse(); why != "" {
		t.Errorf("the landscape table was refused: %s", why)
	}
	marks := CellStats{Tables: 1, Bands: 7, Rows: 7, Columns: 17, Cells: 129, WordsPlaced: 81, WordsRead: 95}
	if why := marks.refuse(); !strings.Contains(why, "95 words") {
		t.Errorf("a mark grid's refusal should count the 95 words read: %q", why)
	}
}

func TestASpreadWordIsWiderPerCharacterThanPrintCanBe(t *testing.T) {
	for _, tc := range []struct {
		name string
		w    tsvWord
		want bool
	}{
		// From the dashed matrix: a dash rule 225 px long read as two letters 7 px tall.
		{"a rule read as ee", tsvWord{w: 225, h: 7, text: "ee"}, true},
		// The widest confident word over the whole corpus, both passes: 1.81 heights per character.
		{"the widest print measured", tsvWord{w: 58, h: 16, text: "on"}, false},
		// A dash is meant to be wide; punctuation is never judged.
		{"an em dash", tsvWord{w: 30, h: 7, text: "—"}, false},
		{"a row of dashes", tsvWord{w: 400, h: 6, text: "-----"}, false},
		{"exactly at the limit", tsvWord{w: 50, h: 10, text: "ab"}, true},
		{"no height", tsvWord{w: 50, h: 0, text: "ab"}, false},
	} {
		if got := tc.w.spread(); got != tc.want {
			t.Errorf("%s: spread = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestALineGoesOnlyWhenItsSpreadWordsCarryIt(t *testing.T) {
	sure := func(s string) tsvWord { return tsvWord{w: 20 * len(s), h: 30, conf: 95, text: s} }
	unsure := func(s string) tsvWord { return tsvWord{w: 10 * len(s), h: 28, conf: 30, text: s} }
	wide := tsvWord{w: 200, h: 7, conf: 30, text: "ee"}
	for _, tc := range []struct {
		name string
		line []tsvWord
		want bool
	}{
		{"print with one broken word", []tsvWord{sure("Machine"), sure("Tool"), wide}, false},
		{"mostly rule", []tsvWord{wide, wide, sure("a")}, true},
		// The narrow pieces where a vertical dash crosses the rule read "ce", unsure and not spread.
		{"rule and its crossings", []tsvWord{wide, unsure("ce"), unsure("ce"), unsure("ce")}, true},
		{"unsure print, nothing spread", []tsvWord{unsure("Grinding"), unsure("etc")}, false},
		{"punctuation only", []tsvWord{{w: 300, h: 6, text: "----"}}, false},
	} {
		if got := spreadLine(tc.line); got != tc.want {
			t.Errorf("%s: removed = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A WORD'S PLACE IN A CELL DOES NOT DEPEND ON WHICH OTHER WORDS ARE THERE. Both cases are the
// corpus's: a 276 px box read off an emblem beside "LIBERTY LOAN COMMITTEE" (loc-yiddish-mixedscript),
// and a 7 px speck inside "April" that tesseract read on the line above (nbs602-form).
func TestReadingLinesKeepsLayoutLinesWhole(t *testing.T) {
	word := func(line, x, y, w, h int, s string) tsvWord {
		return tsvWord{x: x, y: y, w: w, h: h, conf: 96, text: s, line: lineKey{1, 1, line}}
	}
	liberty := []tsvWord{
		word(1, 3893, 6647, 266, 47, "LIBERTY"), word(1, 4172, 6645, 161, 43, "LOAN"), word(1, 4348, 6634, 353, 50, "COMMITTEE"),
		word(2, 3896, 6711, 164, 30, "Second"), word(2, 4090, 6707, 177, 28, "Federal"),
		word(3, 247, 6717, 210, 276, "a"),
	}
	date := []tsvWord{
		word(1, 200, 200, 281, 35, "BIBLIOGRAPHIC"), word(1, 498, 199, 97, 34, "DATA"), word(1, 2016, 239, 7, 7, ";"),
		word(2, 186, 246, 114, 35, "SHEET"), word(2, 1922, 237, 134, 43, "April"), word(2, 2089, 238, 104, 36, "1981"),
	}
	for _, tc := range []struct {
		name  string
		words []tsvWord
		want  string
	}{
		{"an emblem's box", liberty, "LIBERTY LOAN COMMITTEE Second Federal a"},
		{"a speck inside a word", date, "BIBLIOGRAPHIC DATA ; SHEET April 1981"},
	} {
		got := Cell{words: tc.words}.Text()
		if got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
		// The same words, minus any one, keep their order.
		for i := range tc.words {
			rest := append(append([]tsvWord(nil), tc.words[:i]...), tc.words[i+1:]...)
			wantRest := strings.Join(strings.Fields(strings.Replace(" "+tc.want+" ", " "+tc.words[i].text+" ", " ", 1)), " ")
			if got := (Cell{words: rest}).Text(); got != wantRest {
				t.Errorf("%s without %q: %q, want %q", tc.name, tc.words[i].text, got, wantRest)
			}
		}
	}
}
