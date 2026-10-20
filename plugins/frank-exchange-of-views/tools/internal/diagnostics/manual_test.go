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
	text := manualText("feov-record", nil, []string{"finding"}, []string{"inquiry", "propose"})
	pages := ManualPages(text, "feov-record")
	var got []string
	for _, p := range pages {
		got = append(got, strings.Join(p.Path, " "))
	}
	if want := []string{"", "finding", "inquiry propose"}; !slices.Equal(got, want) {
		t.Fatalf("pages = %q, want %q", got, want)
	}
	// A PAGE CLOSED BY THE NEXT RULE AND THE LAST PAGE are both exactly what was printed — the
	// first is where a boundary could eat the newline, and did.
	for i, want := range map[int]string{1: "help for [finding]\n\nFlags:\n  -h, --help\n", 2: "help for [inquiry propose]\n\nFlags:\n  -h, --help\n"} {
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

// THE FOUR OTHER WAYS A BLOCK COMES BACK, each on the smallest manual that uses it: the closing
// block on a page that carries no marker for it, and not on one that marks it elsewhere; a block
// of several paragraphs; a section's entries under a page's own heading; and a description's
// closing sentences after the ones the flag keeps on its line.
func TestExpandManualPutsBackWhatAPageDoesNotMark(t *testing.T) {
	global := "Global Flags:\n      --json   emit JSON\n      --run string   the run"
	pair := "FREE TEXT. Capture it first:\n\n  X=$(cat <<'EOF'\n  EOF\n  )"
	tail := "Use this rather than the other."
	text := "feov-record manual — lean\n\n" + ManualSharedHeading + "\n" +
		ManualClosingLabel(1, 1) + "\n" + global + "\n\n" +
		ManualSharedLabel(2, 2) + "\n" + pair + "\n\n" +
		ManualSharedLabel(3, 3) + "\n" + tail + "\n\n" +
		ManualRule + "\n" + ManualHeader("feov-record", []string{"a"}) + "\nabout a\n\n" + ManualMarker("(FREE TEXT. Capture it first:) ", 2) + "\n\nFlags:\n" + ManualMarker("      --quote text   select only the edits. ", 3) + "\n" +
		ManualRule + "\n" + ManualHeader("feov-record", nil) + "\nthe surface\n\nFlags:\n" + ManualMarker(ManualEntriesLead, 1) + "\n  -v, --version   version\n\nUse it.\n" +
		ManualRule + "\n" + ManualHeader("feov-record", []string{"b"}) + "\nabout b\n\n" + ManualMarker("(Global Flags:) ", 1) + "\n\nUse b.\n"
	pages, err := ExpandManual(text, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"about a\n\n" + pair + "\n\nFlags:\n      --quote text   select only the edits. " + tail + "\n\n" + global + "\n",
		"the surface\n\nFlags:\n      --json   emit JSON\n      --run string   the run\n  -v, --version   version\n\nUse it.\n",
		"about b\n\n" + global + "\n\nUse b.\n",
	}
	for i, p := range pages {
		if p.Body != want[i] {
			t.Errorf("page %d expands to\n%q\nwant\n%q", i, p.Body, want[i])
		}
	}
}

// WHAT A PAGE LEAVES OFF COMES BACK EXACTLY, AND ONLY WHAT WAS LEFT OFF. The help flag returns at
// its sorted place and its section's column, or as a section of its own; the Usage block returns
// before the Flags section; and a page whose shape the restoration would get wrong keeps its lines.
func TestALeanPageRestoresToTheHelpItCameFrom(t *testing.T) {
	const bin = "feov-record"
	global := "\nGlobal Flags:\n      --json   emit JSON\n"
	for name, c := range map[string]struct {
		path       []string
		help, lean string
	}{
		"the help flag between two others": {
			[]string{"verify"},
			"judge it\n\nUsage:\n  feov-record verify [flags]\n\nFlags:\n      --anchor string   which\n  -h, --help            help for verify\n      --reason string   why\n" + global,
			"judge it\n\nFlags:\n      --anchor string   which\n      --reason string   why\n" + global,
		},
		"the help flag last": {
			[]string{"inquiry", "propose"},
			"propose it\n\nUsage:\n  feov-record inquiry propose [flags]\n\nFlags:\n      --anchor string   what\n  -h, --help            help for propose\n" + global,
			"propose it\n\nFlags:\n      --anchor string   what\n" + global,
		},
		"no flag but help": {
			[]string{"show", "board"},
			"read it\n\nUsage:\n  feov-record show board [flags]\n\nFlags:\n  -h, --help   help for board\n" + global,
			"read it\n" + global,
		},
		"no flag but help, and nothing after it": {
			[]string{"assemble"},
			"build it\n\nUsage:\n  feov-record assemble [flags]\n\nFlags:\n  -h, --help   help for assemble\n",
			"build it\n",
		},
		"the surface's own page keeps a Usage that says more": {
			nil,
			"the surface\n\nUsage:\n  feov-record [flags]\n  feov-record [command]\n\nFlags:\n  -h, --help      help for feov-record\n      --json      emit JSON\n  -v, --version   version for feov-record\n",
			"the surface\n\nUsage:\n  feov-record [flags]\n  feov-record [command]\n\nFlags:\n      --json      emit JSON\n  -v, --version   version for feov-record\n",
		},
		// Restored before the Flags section, the Usage would land after the aliases it precedes:
		// the page keeps it, and gives up only the help flag.
		"a Usage that is not directly before the flags stays": {
			[]string{"fetch"},
			"read it\n\nUsage:\n  feov-record fetch [flags]\n\nAliases:\n  fetch, get\n\nFlags:\n  -h, --help         help for fetch\n      --url string   where\n" + global,
			"read it\n\nUsage:\n  feov-record fetch [flags]\n\nAliases:\n  fetch, get\n\nFlags:\n      --url string   where\n" + global,
		},
	} {
		lean := LeanManualPage(bin, c.path, c.help)
		if lean != c.lean {
			t.Errorf("%s: the lean page is\n%q\nwant\n%q", name, lean, c.lean)
		}
		if got := RestoreManualPage(bin, c.path, lean); got != c.help {
			t.Errorf("%s: the lean page restores to\n%q\nwant\n%q", name, got, c.help)
		}
	}
	// A page with no help flag at all cannot be told from one that left it off: it is handed back
	// whole, and it does not restore to itself — which is what the writer's read-back refuses.
	const odd = "no flags here\n"
	if lean := LeanManualPage(bin, []string{"odd"}, odd); lean != odd || RestoreManualPage(bin, []string{"odd"}, lean) == odd {
		t.Errorf("a page that prints no help flag came back lean as %q, or restores to itself", lean)
	}
}

// THE LICENCE IS THE LINE: a manual that does not say it left the Usage and the help flag off is
// read as printed, and one that says so has them put back — except on a page that did not run.
func TestExpandManualRestoresOnlyWhereTheManualSaysItLeftOff(t *testing.T) {
	page := func(p string) string {
		return ManualRule + "\n" + ManualHeader("feov-record", []string{p}) + "\nabout " + p + "\n\nFlags:\n      --reason string   why\n"
	}
	failed := ManualRule + "\n" + ManualHeader("feov-record", []string{"c"}) + "\n" + ManualFailedLead + ", so what follows is its error: boom\npartial\n"
	asPrinted, err := ExpandManual("feov-record manual\n\n"+page("a"), "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	if want := "about a\n\nFlags:\n      --reason string   why\n"; asPrinted[0].Body != want {
		t.Errorf("a manual that left nothing off was changed in the reading: %q", asPrinted[0].Body)
	}
	restored, err := ExpandManual("feov-record manual\n\n"+ManualLeftOff+"\n\n"+page("a")+failed, "feov-record")
	if err != nil {
		t.Fatal(err)
	}
	if want := "about a\n\nUsage:\n  feov-record a [flags]\n\nFlags:\n  -h, --help            help for a\n      --reason string   why\n"; restored[0].Body != want {
		t.Errorf("the page restores to %q, want %q", restored[0].Body, want)
	}
	if want := ManualFailedLead + ", so what follows is its error: boom\npartial\n"; restored[1].Body != want {
		t.Errorf("a page that did not run was given lines it never printed: %q", restored[1].Body)
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

// THE READER ACCEPTS WHAT THE WRITER PRINTS, AND NOTHING NEAR IT. A label or marker is read with
// the writer's own format and accepted only where the writer prints the line back, so the two
// cannot drift apart: a reworded label changes what both sides speak.
func TestSharedLabelsAndMarkersAreReadWithTheWritersFormat(t *testing.T) {
	for _, c := range []struct{ id, pages int }{{1, 2}, {7, 31}, {12, 140}} {
		if id, closes, ok := sharedLabel(ManualSharedLabel(c.id, c.pages)); !ok || id != c.id || closes {
			t.Errorf("ManualSharedLabel(%d, %d) read as (%d, closing=%v, %v)", c.id, c.pages, id, closes, ok)
		}
		if id, closes, ok := sharedLabel(ManualClosingLabel(c.id, c.pages)); !ok || id != c.id || !closes {
			t.Errorf("ManualClosingLabel(%d, %d) read as (%d, closing=%v, %v)", c.id, c.pages, id, closes, ok)
		}
		lead := "      --reason string   "
		if id, got, ok := markedBlock(ManualMarker(lead, c.id)); !ok || id != c.id || got != lead {
			t.Errorf("ManualMarker(%q, %d) read as (%d, %q, %v)", lead, c.id, id, got, ok)
		}
	}
	for _, l := range []string{
		"§3 (the last block of 4 pages — some other words):",
		"§03 (on 2 pages):",
		"§3 (on 2 pages): and more",
	} {
		if _, _, ok := sharedLabel(l); ok {
			t.Errorf("%q was read as a SHARED label; the writer prints no such line", l)
		}
	}
	for _, l := range []string{"see → SHARED §03", "see → SHARED §+3", "see → SHARED §3 "} {
		if _, _, ok := markedBlock(l); ok {
			t.Errorf("%q was read as a marker; the writer prints no such line", l)
		}
	}
}
