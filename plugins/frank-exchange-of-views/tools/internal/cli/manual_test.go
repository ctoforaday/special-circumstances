package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// In this test binary the executor builds each page in-process (see selfExecHelp), so these tests
// assert real help text; the self-exec path is driven by the difftest scenario, against the built
// binary.

// manualOf runs manual on a seat's surface and returns what it printed and its pages with every
// SHARED block put back — read with the survey's own parser, so the writer and the reader are held
// to one grammar.
func manualOf(t *testing.T, seatID string) (string, []diagnostics.ManualPage) {
	t.Helper()
	out, err := run(t, manualName, "--seat-id", seatID)
	if err != nil {
		t.Fatalf("manual on %s: %v\n%s", seatID, err, out)
	}
	pages, err := diagnostics.ExpandManual(out, InvokedAs())
	if err != nil {
		t.Fatalf("manual on %s does not expand: %v", seatID, err)
	}
	return out, pages
}

// surfaceOf is what manual must list for a seat: the surface's own page, then every command in tree
// order that gets a page.
func surfaceOf(seatID string) []string {
	out := []string{""}
	walkSurface(NewRootFor(seatID), func(p []string, c *cobra.Command) {
		if !leftOutOfManual(c) {
			out = append(out, strings.Join(p, " "))
		}
	})
	return out
}

func pathsOf(pages []diagnostics.ManualPage) []string {
	var out []string
	for _, p := range pages {
		out = append(out, strings.Join(p.Path, " "))
	}
	return out
}

func everySurface() map[string]string {
	seats := map[string]string{record.OperatorRole: record.OperatorRole}
	for _, r := range []string{"lens", "merge", "blue", "bench"} {
		seats[r] = record.SampleSeatOf(r)
	}
	return seats
}

// EVERY COMMAND ON THE SURFACE, EVERY SURFACE, IN TREE ORDER — and the opening line says how many,
// and says the pages are lean on purpose.
func TestManualListsEveryCommandOnEachSurface(t *testing.T) {
	for role, seatID := range everySurface() {
		out, pages := manualOf(t, seatID)
		want := surfaceOf(seatID)
		if len(want) < 10 {
			t.Fatalf("%s: the walk found %d commands — too few to be a surface, and this test would pass over nothing", role, len(want))
		}
		if got := pathsOf(pages); !slices.Equal(got, want) {
			t.Errorf("%s: manual listed %d pages, the surface holds %d\n got: %q\nwant: %q", role, len(got), len(want), got, want)
		}
		first, _, _ := strings.Cut(out, "\n")
		for _, w := range []string{"the " + role + " surface", "seat " + seatID, ": " + strconv.Itoa(len(want)-1) + " commands", "LEAN ON PURPOSE", "SHARED"} {
			if !strings.Contains(first, w) {
				t.Errorf("%s: the opening line %q does not say %q", role, first, w)
			}
		}
		// EVERY LEAF HAS A PAGE, whatever the group rule decides — a leaf is a command a seat runs.
		have := map[string]bool{}
		for _, p := range pathsOf(pages) {
			have[p] = true
		}
		walkSurface(NewRootFor(seatID), func(p []string, c *cobra.Command) {
			k := strings.Join(p, " ")
			if !c.HasSubCommands() && k != manualName && !have[k] {
				t.Errorf("%s: no page for the command %q", role, k)
			}
		})
		// AND THE JOIN KEYS AGREE: every command the gates attribute to this role is a page here.
		for _, k := range CommandPaths() {
			if typed, ok := strings.CutPrefix(k, role+" "); ok && !have[typed] {
				t.Errorf("%s: CommandPaths has %q and the manual has no page for %q", role, k, typed)
			}
		}
		if have[manualName] {
			t.Errorf("%s: manual printed its own page — the page a seat is already reading", role)
		}
	}
}

// A GROUP LOSES ITS PAGE ONLY WHEN THE PAGE SAYS NOTHING OF ITS OWN — checked against the page the
// group actually prints, not against the rule that dropped it. Every paragraph of a dropped group's
// help must be its menu line, a heading section (usage, listing, flags) or cobra's closing pointer.
func TestADroppedGroupPageHeldNothingButItsListing(t *testing.T) {
	structural := regexp.MustCompile(`^(Usage:|Aliases:|Available Commands:|Command groups|Flags:|Global Flags:|Use ")`)
	for role, seatID := range everySurface() {
		var kept, dropped []string
		walkSurface(NewRootFor(seatID), func(p []string, c *cobra.Command) {
			if !c.HasSubCommands() {
				return
			}
			k := strings.Join(p, " ")
			if !leftOutOfManual(c) {
				kept = append(kept, k)
				return
			}
			dropped = append(dropped, k)
			help, err := run(t, append(append([]string{}, p...), "--help", "--seat-id", seatID)...)
			if err != nil {
				t.Fatal(err)
			}
			for i, para := range strings.Split(strings.TrimSpace(help), "\n\n") {
				if i == 0 && strings.TrimSpace(para) == strings.TrimSpace(c.Short) {
					continue
				}
				if !structural.MatchString(strings.TrimSpace(para)) {
					t.Errorf("%s: group %q was left out of the manual, and its page carries prose of its own:\n%s", role, k, para)
				}
			}
		})
		t.Logf("%s: group pages kept %q, left out %q", role, kept, dropped)
	}
}

// EVERY PAGE IS THAT COMMAND'S OWN HELP, BYTE FOR BYTE, once its SHARED blocks are put back —
// under a header naming it. A block lifted to SHARED whose marker was dropped fails here.
func TestEveryManualPageIsThatCommandsOwnHelp(t *testing.T) {
	for _, seatID := range []string{record.SampleSeatOf("blue"), record.SampleSeatOf("lens")} {
		out, pages := manualOf(t, seatID)
		if len(pages) < 10 {
			t.Fatalf("%s: %d pages parsed — too few to be a surface", seatID, len(pages))
		}
		for _, p := range pages {
			header := diagnostics.ManualHeader(InvokedAs(), p.Path)
			if !strings.Contains(out, diagnostics.ManualRule+"\n"+header+"\n") {
				t.Errorf("%s: no rule-and-header %q", seatID, header)
			}
			args := append(append([]string{}, p.Path...), "--help", "--seat-id", seatID)
			want, err := run(t, args...)
			if err != nil {
				t.Fatalf("%s: %v", strings.Join(args, " "), err)
			}
			if p.Body != want {
				i := 0
				for i < len(p.Body) && i < len(want) && p.Body[i] == want[i] {
					i++
				}
				from := max(0, i-120)
				t.Errorf("%s: the page under %q is not that command's help (page %d bytes, help %d; they part at byte %d)\n--- page ---\n%.360s\n--- %s ---\n%.360s",
					seatID, header, len(p.Body), len(want), i, p.Body[from:], strings.Join(args, " "), want[from:])
			}
		}
	}
}

var markerID = regexp.MustCompile(`→ SHARED §(\d+)$`)
var sharedLabel = regexp.MustCompile(`^§(\d+) \(on (\d+) pages\):$`)

// THE ECONOMY IS COMPLETE AND HONEST: every block that repeats verbatim on two pages is lifted, each
// lifted block is marked on exactly as many pages as its label claims, and the manual comes out a
// fraction of the pages it stands for.
func TestTheManualLiftsEveryVerbatimRepeatAndMarksEachRemoval(t *testing.T) {
	for role, seatID := range everySurface() {
		out, expanded := manualOf(t, seatID)
		raw := diagnostics.ManualPages(out, InvokedAs())
		claimed := map[string]int{}
		for _, l := range strings.Split(out, "\n") {
			if m := sharedLabel.FindStringSubmatch(l); m != nil {
				claimed[m[1]], _ = strconv.Atoi(m[2])
			}
		}
		if len(claimed) == 0 {
			t.Fatalf("%s: nothing was lifted to SHARED — every page inherits the same global flags, so the economy did not run", role)
		}
		marked := map[string]int{}
		paraPages := map[string]map[int]bool{}
		for i, p := range raw {
			for _, l := range strings.Split(p.Body, "\n") {
				if m := markerID.FindStringSubmatch(l); m != nil {
					marked[m[1]]++
				}
			}
			for _, para := range strings.Split(p.Body, "\n\n") {
				para = strings.Trim(para, "\n")
				if len(para) >= minShared {
					if paraPages[para] == nil {
						paraPages[para] = map[int]bool{}
					}
					paraPages[para][i] = true
				}
			}
		}
		for id, n := range claimed {
			if marked[id] != n || n < 2 {
				t.Errorf("%s: SHARED §%s says it is on %d pages and %d pages mark it", role, id, n, marked[id])
			}
		}
		for para, on := range paraPages {
			if len(on) >= 2 {
				t.Errorf("%s: a %d-char block is still printed on %d pages rather than once under SHARED:\n%.200s", role, len(para), len(on), para)
			}
		}
		whole := 0
		for _, p := range expanded {
			whole += len(p.Body)
		}
		// THE WORDS SECTION IS NOT A PAGE AND STANDS FOR NONE, so it is left out of a measure of
		// how much the lifting saved. Counted in, it reads as lost economy on the smallest surface.
		lifted := len(out) - len(wordsSection(out))
		t.Logf("%s: manual %d chars (%d without the words section) against %d chars of the pages it stands for (%d%%), %d blocks in SHARED", role, len(out), lifted, whole, 100*lifted/whole, len(claimed))
		// MEASURED 63–72% across the five surfaces (the operator's pages share least). A manual
		// whose repeats stopped being lifted comes out at or above 100%, markers and all, so 80%
		// is a bound that fires on the economy failing rather than on a page gaining a sentence.
		if lifted > whole*8/10 {
			t.Errorf("%s: the manual is %d chars against %d for the pages it stands for — the repeats are no longer being lifted", role, lifted, whole)
		}
	}
}

// NO LINE OUTRUNS A READ. The prompt sends a seat to read the manual from a file, and the Read tool
// truncates any line past 2,000 characters — so a long line is the tail of a sentence the seat never
// sees, with nothing on the page to say it was cut.
func TestNoManualLineOutrunsARead(t *testing.T) {
	const readLineCap = 2000
	for role, seatID := range everySurface() {
		out, _ := manualOf(t, seatID)
		for i, l := range strings.Split(out, "\n") {
			if n := len([]rune(l)); n > readLineCap {
				t.Errorf("%s: line %d of the manual is %d characters, past the %d a Read returns: %.120s…", role, i+1, n, readLineCap, l)
			}
		}
	}
}

// A SEAT'S MANUAL IS ITS OWN SURFACE AND NOBODY ELSE'S.
func TestManualNamesOnlyThisSeatsCommands(t *testing.T) {
	cases := []struct{ seat, has, hasNot string }{
		{record.SampleSeatOf("blue"), "line-of-inquiry propose", "mint"},
		{record.SampleSeatOf("lens"), "mint", "line-of-inquiry propose"},
	}
	for _, c := range cases {
		_, pages := manualOf(t, c.seat)
		have := map[string]bool{}
		for _, p := range pathsOf(pages) {
			have[p] = true
		}
		if !have[c.has] {
			t.Errorf("%s: no page for %q", c.seat, c.has)
		}
		if have[c.hasNot] {
			t.Errorf("%s: a page for %q, which is another seat's command", c.seat, c.hasNot)
		}
		own := map[string]bool{}
		for _, p := range surfaceOf(c.seat) {
			own[p] = true
		}
		for p := range have {
			if !own[p] {
				t.Errorf("%s: a page for %q, which is not on this seat's tree", c.seat, p)
			}
		}
	}
}

// A PAGE THAT FAILS IS LISTED WITH ITS ERROR, AND THE COMMAND SAYS SO.
func TestAManualPageThatFailsIsReportedNotSkipped(t *testing.T) {
	orig := runHelp
	t.Cleanup(func() { runHelp = orig })
	runHelp = func(argv []string) (string, error) {
		if argv[0] == "cite" {
			return "partial output\n", errors.New("the page exploded")
		}
		return orig(argv)
	}
	seatID := record.SampleSeatOf("blue")
	out, err := run(t, manualName, "--seat-id", seatID)
	if err == nil || !strings.Contains(err.Error(), "1 of") || !strings.Contains(err.Error(), "cite") {
		t.Fatalf("manual with one failing page returned %v — it must fail, naming the page", err)
	}
	pages, xerr := diagnostics.ExpandManual(out, InvokedAs())
	if xerr != nil {
		t.Fatal(xerr)
	}
	if got, want := len(pages), len(surfaceOf(seatID)); got != want {
		t.Errorf("%d pages with one failing, want all %d — a failed page was skipped", got, want)
	}
	for _, p := range pages {
		if strings.Join(p.Path, " ") == "cite" {
			if !strings.Contains(p.Body, "DID NOT RUN") || !strings.Contains(p.Body, "the page exploded") || !strings.Contains(p.Body, "partial output") {
				t.Errorf("the failing page does not carry its error and output:\n%s", p.Body)
			}
			return
		}
	}
	t.Error("no page for cite at all")
}

func TestManualRefusesJSONRatherThanIgnoringIt(t *testing.T) {
	if _, err := run(t, manualName, "--seat-id", record.SampleSeatOf("bench"), "--json"); err == nil {
		t.Error("manual --json succeeded — a flag that changes nothing must be refused, not dropped")
	}
}

// readResult renders lines the way the Read tool returns them: a right-aligned line number and a
// tab on every line.
func readResult(lines []string, from int) string {
	var b strings.Builder
	for i, l := range lines {
		fmt.Fprintf(&b, "%6d\t%s\n", from+i, l)
	}
	return b.String()
}

func writeTrajectory(t *testing.T, steps ...map[string]any) string {
	t.Helper()
	var b []byte
	for _, s := range steps {
		j, _ := json.Marshal(map[string]any{"message": map[string]any{"content": []any{s}}})
		b = append(append(b, j...), '\n')
	}
	p := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// THE SURVEY CREDITS THE MANUAL THE PROMPT PRESCRIBES — sent to a file and read back, in windows,
// from real output. The Bash call's own result is EMPTY in that shape, so a survey that looked only
// there scored every later command as run blind.
func TestTheSurveyReadsARedirectedManualThroughItsRead(t *testing.T) {
	seatID := record.SampleSeatOf("blue")
	out, _ := manualOf(t, seatID)
	bin := "/tmp/x/" + InvokedAs()
	file := "/scratch/s1/manual.txt"
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	half := len(lines) / 2
	use := func(id, name string, input map[string]any) map[string]any {
		return map[string]any{"type": "tool_use", "name": name, "id": id, "input": input}
	}
	res := func(id, text string) map[string]any {
		return map[string]any{"type": "tool_result", "tool_use_id": id, "content": text}
	}
	steps := []map[string]any{
		use("a", "Bash", map[string]any{"command": `"` + bin + `" --seat-id ` + seatID + ` manual > ` + file}), res("a", ""),
		use("r1", "Read", map[string]any{"file_path": file, "limit": half}), res("r1", readResult(lines[:half], 1)),
		use("r2", "Read", map[string]any{"file_path": file, "offset": half + 1}), res("r2", readResult(lines[half:], half+1)),
		use("b", "Bash", map[string]any{"command": `"` + bin + `" line-of-inquiry propose --reason x`}), res("b", "ok"),
	}
	s, err := diagnostics.ReadSurvey(writeTrajectory(t, steps...), InvokedAs(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.FirstUses) != 1 || s.FirstUses[0].Command != "line-of-inquiry propose" || s.FirstUses[0].Depth != diagnostics.DepthCommand {
		t.Errorf("first uses = %+v, want line-of-inquiry propose at depth command — its page was read from the file", s.FirstUses)
	}
	if got, want := len(s.HelpPages), len(surfaceOf(seatID)); got != want || s.ManualUnread != 0 {
		t.Errorf("survey saw %d pages (unread %d), the manual printed %d", got, s.ManualUnread, want)
	}
	if s.TotalCalls != 2 || s.FirstUses[0].Call != 2 {
		t.Errorf("total calls %d, first use at call %d — a Read is not an invocation of the tool", s.TotalCalls, s.FirstUses[0].Call)
	}

	// AND A MANUAL SENT TO A FILE NOBODY READ WAS NOT SHOWN, and says so.
	unread, err := diagnostics.ReadSurvey(writeTrajectory(t, steps[0], steps[1], steps[6], steps[7]), InvokedAs(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if unread.ManualUnread != 1 || unread.FirstUses[0].Depth == diagnostics.DepthCommand {
		t.Errorf("an unread redirected manual: unread=%d depth=%q, want 1 and no page credit", unread.ManualUnread, unread.FirstUses[0].Depth)
	}
}
