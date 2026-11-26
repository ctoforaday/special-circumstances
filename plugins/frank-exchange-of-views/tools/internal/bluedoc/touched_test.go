package bluedoc

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
)

// AN EDIT STAMPS THE WHOLE PARAGRAPHS IT TOUCHED, before and after, anchors on them and the breaks
// between them kept — every sentence of each, the ones it left alone included.
func TestTouchedParagraphsAreWholeBlocksAroundTheSpan(t *testing.T) {
	const g = "<!--gap:G-000000c3-->"
	before := "# H\n\nFirst sentence. Second one" + g + ".\n\n- item one\n- item two\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\nTail paragraph.\n"
	cases := []struct {
		name, old, new   string
		wantOld, wantNew string
	}{
		{"inside one paragraph", "Second one", "Second two",
			"First sentence. Second one" + g + ".", "First sentence. Second two" + g + "."},
		{"across a paragraph break", "Second one" + g + ".\n\n- item one", "Merged" + g + ".\n\n- item uno",
			"First sentence. Second one" + g + ".\n\n- item one", "First sentence. Merged" + g + ".\n\n- item uno"},
		{"a table line", "| 1 | 2 |", "| 1 | 3 |", "| 1 | 2 |", "| 1 | 3 |"},
		{"a split", "Tail paragraph.", "Tail.\n\nNew paragraph.", "Tail paragraph.", "Tail.\n\nNew paragraph."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start := strings.Index(before, c.old)
			if start < 0 {
				t.Fatalf("%q not in the report", c.old)
			}
			end := start + len(c.old)
			after := before[:start] + c.new + before[end:]
			o, n := TouchedParagraphs(before, after, start, end)
			if o != c.wantOld || n != c.wantNew {
				t.Fatalf("got\n old %q\n new %q\nwant\n old %q\n new %q", o, n, c.wantOld, c.wantNew)
			}
			if !strings.Contains(anchortext.Visible(n), anchortext.Visible(c.new)) {
				t.Fatalf("the new run %q does not hold the replacement %q", n, c.new)
			}
		})
	}
}

// AN ANCHOR THE EDIT LEAVES BARE IS REOPENED: its sentence is gone, so its referent changed. One
// whose own sentence is unchanged in a touched paragraph is not.
func TestReopenedAnchorsListsAnAnchorLeftBare(t *testing.T) {
	const g, c = "<!--gap:G-000000c3-->", "<!--cite:C-000000b2-->"
	before := "Kept sentence" + c + ". Cut sentence" + g + "."
	after := "Kept sentence" + c + ". " + g
	got := ReopenedAnchors(before, after)
	if !slices.Equal(got, []string{"G-000000c3"}) {
		t.Fatalf("an anchor left bare: want [G-000000c3], got %v", got)
	}
	// Already bare on both sides: the edit did not bare it.
	if got := ReopenedAnchors(after, after+" More."); len(got) != 0 {
		t.Fatalf("an anchor bare before the edit: want none, got %v", got)
	}
}
