package reportproj

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE REPORT IS THE PARTS IN EVENT ORDER. Each part after the first follows the text before it, its
// trailing blank lines trimmed, after one blank line; a change folds over the joined text.
func TestThePartsJoinInTheOrderTheyAreRecorded(t *testing.T) {
	ops := []record.ReportOp{
		{EventID: 3, Kind: "append", A: "## Lane two\n\nSecond.\n\n\n"},
		{EventID: 5, Kind: "append", A: "## Lane one\n\nFirst.\n"},
		{EventID: 8, Kind: "edit", A: "Second.", B: "Second, edited."},
		{EventID: 9, Kind: "append", A: "## Lane three\n\nThird.\n"},
	}
	got, err := Replay("", false, ops)
	if err != nil {
		t.Fatal(err)
	}
	want := "## Lane two\n\nSecond, edited.\n\n## Lane one\n\nFirst.\n\n## Lane three\n\nThird.\n"
	if got != want {
		t.Errorf("the parts joined as\n%q\nwant\n%q", got, want)
	}
}

func TestJoinPart(t *testing.T) {
	for _, c := range []struct{ text, part, want string }{
		{"", "A.\n", "A.\n"},
		{"A.\n", "B.\n", "A.\n\nB.\n"},
		{"A.\n\n\n", "B.\n", "A.\n\nB.\n"},
		{"A.", "B.", "A.\n\nB."},
	} {
		if got := JoinPart(c.text, c.part); got != c.want {
			t.Errorf("JoinPart(%q, %q) = %q, want %q", c.text, c.part, got, c.want)
		}
	}
}

// A REPORT NO LANE HAS RECORDED A PART OF IS NOTHING TO READ OR CHANGE: the empty fold, and a change
// before any part, are each NoPart — never an empty report rendered as if it were one.
func TestAReportWithNoPartIsRefused(t *testing.T) {
	for name, ops := range map[string][]record.ReportOp{
		"the empty fold":          nil,
		"an edit before a part":   {{Kind: "edit", A: "x", B: "y"}, {Kind: "append", A: "x\n"}},
		"an insert before a part": {{Kind: "insert", A: "x", B: "F-00000001"}},
		"a remove before a part":  {{Kind: "remove", A: "F-00000001"}},
	} {
		if _, err := Replay("", false, ops); err == nil || err.Error() != NoPart {
			t.Errorf("%s: replayed with %v, want the refusal %q", name, err, NoPart)
		}
	}
	if got, err := Replay("A.\n", true, nil); err != nil || got != "A.\n" {
		t.Errorf("a fold that starts on a part replays nothing as (%q, %v)", got, err)
	}
}
