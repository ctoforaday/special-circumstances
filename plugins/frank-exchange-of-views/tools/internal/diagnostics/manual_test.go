package diagnostics

import (
	"slices"
	"strings"
	"testing"
)

func manualText(bin string, pages ...[]string) string {
	var b strings.Builder
	b.WriteString(bin + " manual — the lens surface, seat red-lens-evidence: 2 commands\n")
	for _, p := range pages {
		b.WriteString(ManualRule + "\n" + ManualHeader(bin, p) + "\n")
		b.WriteString("help for [" + strings.Join(p, " ") + "]\n\nFlags:\n  -h, --help\n")
	}
	return b.String()
}

// THE HEADER ROUND-TRIPS, and a rule that is not followed by this tool's header is body text.
func TestManualPagesRoundTripTheHeader(t *testing.T) {
	text := manualText("feov-record", nil, []string{"finding"}, []string{"class", "new"})
	pages := ManualPages(text, "feov-record")
	var got []string
	for _, p := range pages {
		got = append(got, strings.Join(p.Path, " "))
	}
	if want := []string{"", "finding", "class new"}; !slices.Equal(got, want) {
		t.Fatalf("pages = %q, want %q", got, want)
	}
	// A PAGE CLOSED BY THE NEXT RULE AND THE LAST PAGE are both exactly what was printed — the
	// first is where a boundary could eat the newline, and did.
	for i, want := range map[int]string{1: "help for [finding]\n\nFlags:\n  -h, --help\n", 2: "help for [class new]\n\nFlags:\n  -h, --help\n"} {
		if pages[i].Body != want {
			t.Errorf("page %d body = %q, want %q", i, pages[i].Body, want)
		}
	}
	if n := len(ManualPages(text, "other-tool")); n != 0 {
		t.Errorf("%d pages recognised under another tool's name", n)
	}
	if n := len(ManualPages(ManualRule+"\n$ feov-record finding\n", "feov-record")); n != 0 {
		t.Errorf("a header without --help was taken for a page")
	}
}

// A SEAT THAT READ ITS MANUAL HAD EVERY COMMAND'S OWN PAGE, and the survey says so.
func TestAManualIsAHelpReadOfEveryPage(t *testing.T) {
	p := traj(t,
		use("a", `"/tmp/x/feov-record" --seat-id red-lens-evidence manual`),
		res("a", manualText("feov-record", nil, []string{"finding"}, []string{"verify"}), false),
		use("b", `"/tmp/x/feov-record" finding --quote q`), res("b", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", map[string]bool{"finding": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.FirstUses) != 1 || s.FirstUses[0].Command != "finding" {
		t.Fatalf("first uses = %+v — the manual call must not count as a command run", s.FirstUses)
	}
	if s.FirstUses[0].Depth != DepthCommand {
		t.Errorf("depth = %q, want %q — the manual printed finding's own page", s.FirstUses[0].Depth, DepthCommand)
	}
	if len(s.HelpPages) != 3 || s.LastHelpCall != 1 {
		t.Errorf("help pages = %q (last help call %d), want the three the manual printed", s.HelpPages, s.LastHelpCall)
	}
	if s.Traversal.Unclassified != 0 || len(s.Traversal.Sequence) != 1 {
		t.Errorf("traversal = %+v — reading the manual is not an operation on the surface", s.Traversal)
	}
}

// A MANUAL THE SURVEY COULD NOT READ IS COUNTED AS SUCH, never as a seat shown nothing.
func TestAManualWithNoReadablePageIsCounted(t *testing.T) {
	p := traj(t,
		use("a", `"/tmp/x/feov-record" manual`), res("a", "feov-record: something went wrong", true),
		use("b", `"/tmp/x/feov-record" finding`), res("b", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.ManualUnread != 1 {
		t.Errorf("manualUnread = %d, want 1", s.ManualUnread)
	}
	if s.FirstUses[0].Depth != DepthNone {
		t.Errorf("depth = %q — an unread manual supplied no page", s.FirstUses[0].Depth)
	}
}
