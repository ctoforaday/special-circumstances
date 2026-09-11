package diagnostics

import (
	"fmt"
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

// EVERY MARKER PUTS ITS BLOCK BACK, and a marker naming a block SHARED does not hold is an error —
// that page has lost a line.
func TestExpandManualPutsEveryMarkerBack(t *testing.T) {
	global := "Global Flags:\n      --json   emit JSON\n      --run string   the run"
	reason := "your THINKING for this act"
	// The same --reason description at two paddings: one lifted description, each page keeping
	// its own padded flag name.
	text := "feov-record manual — lean\n\n" + ManualSharedHeading + "\n" +
		ManualSharedLabel(1, 2) + "\n" + global + "\n\n" +
		ManualSharedLabel(2, 2) + "\n" + reason + "\n\n" +
		ManualRule + "\n" + ManualHeader("feov-record", []string{"a"}) + "\nabout a\n\nFlags:\n" + ManualMarker("      --reason string   ", 2) + "\n\n" + ManualMarker("(Global Flags:) ", 1) + "\n" +
		ManualRule + "\n" + ManualHeader("feov-record", []string{"b"}) + "\nabout b\n\nFlags:\n" + ManualMarker("      --reason string        ", 2) + "\n\n" + ManualMarker("(Global Flags:) ", 1) + "\n"
	pages, err := ExpandManual(text, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	pad := map[string]string{"a": "   ", "b": "        "}
	for _, p := range pages {
		want := "about " + p.Path[0] + "\n\nFlags:\n      --reason string" + pad[p.Path[0]] + reason + "\n\n" + global + "\n"
		if p.Body != want {
			t.Errorf("page %s expands to %q, want %q", p.Path[0], p.Body, want)
		}
	}
	if _, err := ExpandManual(strings.Replace(text, "§1 (on", "§9 (on", 1), "feov-record"); err == nil {
		t.Error("a marker naming no SHARED block expanded without complaint")
	}
}

func TestRedirectTarget(t *testing.T) {
	for cmd, want := range map[string]string{
		`"/b/feov-record" manual > /s/m.txt`:       "/s/m.txt",
		`"/b/feov-record" manual >/s/m.txt 2>&1`:   "/s/m.txt",
		`/b/feov-record manual 2>&1 >> "$S/m.txt"`: "$S/m.txt",
		`/b/feov-record manual 2> /s/err.txt`:      "",
		`/b/feov-record manual | head -50`:         "",
		`/b/feov-record manual &> /s/both.txt`:     "/s/both.txt",
	} {
		if got := redirectTarget(cmd); got != want {
			t.Errorf("redirectTarget(%q) = %q, want %q", cmd, got, want)
		}
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

func readUse(id, path string) any {
	return map[string]any{"message": map[string]any{"content": []any{
		map[string]any{"type": "tool_use", "name": "Read", "id": id, "input": map[string]any{"file_path": path}}}}}
}

func numbered(text string) string {
	var b strings.Builder
	for i, l := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		fmt.Fprintf(&b, "%6d\t%s\n", i+1, l)
	}
	return b.String()
}

// A REDIRECTED MANUAL IS CREDITED WHERE IT IS READ — its Bash result is empty, the pages arrive at
// the Read, and a target the shell had to expand is settled by the pages themselves.
func TestARedirectedManualIsCreditedAtItsRead(t *testing.T) {
	manual := manualText("feov-record", nil, []string{"finding"})
	for _, target := range []string{"/s/manual.txt", "$S/manual.txt"} {
		p := traj(t,
			use("a", `"/tmp/x/feov-record" manual > `+target), res("a", "", false),
			readUse("r", "/s/manual.txt"), res("r", numbered(manual), false),
			use("b", `"/tmp/x/feov-record" finding`), res("b", "ok", false),
		)
		s, err := ReadSurvey(p, "feov-record", nil)
		if err != nil {
			t.Fatal(err)
		}
		if s.ManualUnread != 0 || s.FirstUses[0].Depth != DepthCommand || s.TotalCalls != 2 {
			t.Errorf("target %s: unread=%d depth=%q calls=%d, want 0, command, 2", target, s.ManualUnread, s.FirstUses[0].Depth, s.TotalCalls)
		}
	}
	// A Read of some OTHER file does not settle a literal target.
	p := traj(t,
		use("a", `"/tmp/x/feov-record" manual > /s/manual.txt`), res("a", "", false),
		readUse("r", "/s/notes.md"), res("r", numbered("just notes\n"), false),
		use("b", `"/tmp/x/feov-record" finding`), res("b", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.ManualUnread != 1 {
		t.Errorf("unread = %d, want 1 — the manual's file was never read", s.ManualUnread)
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
