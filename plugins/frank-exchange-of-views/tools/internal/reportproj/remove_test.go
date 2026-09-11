package reportproj

import (
	"strings"
	"testing"
)

// A RETIRED ANCHOR TAKES ITS HUSK WITH IT. The shapes an edit-to-the-bare-anchor leaves, and what
// the report must read as once the retire has taken the anchor out.
func TestRemoveAnchorTakesTheHusk(t *testing.T) {
	const c = "<!--cite:c-1-->"
	for _, tc := range []struct{ name, in, want string }{
		{"an emptied bullet goes whole", "- a.\n- " + "<!--fx:f-1-->" + "?\n- c.\n", "- a.\n- c.\n"},
		{"a paragraph that was only the anchor goes, blank run collapsed", "P one.\n\n" + c + "\n\nP two.\n", "P one.\n\nP two.\n"},
		{"the last paragraph goes with no trailing blank", "P one.\n\n" + c + ".\n", "P one.\n"},
		{"the first line goes with no leading blank", c + "\n\nP two.\n", "P two.\n"},
		{"an emptied mid-line segment goes with its terminator", "One. " + c + ". Two.\n", "One. Two.\n"},
		{"an emptied final segment goes with its space", "One. " + c + "\n", "One.\n"},
		{"at a line start, the seam space goes", c + " Two.\n", "Two.\n"},
		{"at a line start with a stray terminator", c + ". Two.\n", "Two.\n"},
		{"a second anchor keeps the segment until it too leaves", "One. " + c + "<!--fx:f-2-->. Two.\n", "One. <!--fx:f-2-->. Two.\n"},
		{"prose around the anchor is untouched", "One " + c + " two.\n", "One two.\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok := c
			if !strings.Contains(tc.in, c) {
				tok = "<!--fx:f-1-->"
			}
			got := RemoveAnchorAt(tc.in, strings.Index(tc.in, tok), len(tok))
			if got != tc.want {
				t.Errorf("RemoveAnchorAt(%q)\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

// ORDER-INDEPENDENT across several anchors exiting one segment: only the last finds it empty.
func TestRemoveAnchorOrderIndependent(t *testing.T) {
	in := "- a.\n- <!--cite:c-1--><!--fx:f-2-->\n- c.\n"
	rm := func(s, tok string) string { return RemoveAnchorAt(s, strings.Index(s, tok), len(tok)) }
	ab := rm(rm(in, "<!--cite:c-1-->"), "<!--fx:f-2-->")
	ba := rm(rm(in, "<!--fx:f-2-->"), "<!--cite:c-1-->")
	if ab != "- a.\n- c.\n" || ab != ba {
		t.Errorf("order changed the result: %q vs %q", ab, ba)
	}
}

// A remove op naming an anchor the report does not hold is LOUD: the retire validated presence
// at the write, so absence means the record no longer describes a real sequence.
func TestRemoveOfAnAbsentAnchorIsLoud(t *testing.T) {
	if _, err := (removeMut{id: "c-9"}).apply("no anchor here.\n"); err == nil {
		t.Fatal("removing an absent anchor succeeded silently — replay would render a report that never existed")
	}
}
