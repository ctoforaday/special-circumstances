package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

// manualOf runs manual on a seat's surface and splits it with the parser the survey reads it with,
// so the writer and the reader are held to one grammar.
func manualOf(t *testing.T, seatID string) (string, []diagnostics.ManualPage) {
	t.Helper()
	out, err := run(t, manualName, "--seat-id", seatID)
	if err != nil {
		t.Fatalf("manual on %s: %v\n%s", seatID, err, out)
	}
	return out, diagnostics.ManualPages(out, InvokedAs())
}

// surfaceOf is what manual must list for a seat, from the one walk — the surface's own page, then
// every command in tree order.
func surfaceOf(seatID string) []string {
	out := []string{""}
	walkSurface(NewRootFor(seatID), func(p []string, _ *cobra.Command) { out = append(out, strings.Join(p, " ")) })
	return out
}

func pathsOf(pages []diagnostics.ManualPage) []string {
	var out []string
	for _, p := range pages {
		out = append(out, strings.Join(p.Path, " "))
	}
	return out
}

// EVERY COMMAND ON THE SURFACE, EVERY SURFACE, IN TREE ORDER — and the opening line says how many.
func TestManualListsEveryCommandOnEachSurface(t *testing.T) {
	seats := map[string]string{record.OperatorRole: record.OperatorRole}
	for _, r := range []string{"lens", "merge", "blue", "bench"} {
		seats[r] = record.SampleSeatOf(r)
	}
	for role, seatID := range seats {
		out, pages := manualOf(t, seatID)
		want := surfaceOf(seatID)
		if len(want) < 10 {
			t.Fatalf("%s: the walk found %d commands — too few to be a surface, and this test would pass over nothing", role, len(want))
		}
		if got := pathsOf(pages); !slices.Equal(got, want) {
			t.Errorf("%s: manual listed %d pages, the surface holds %d\n got: %q\nwant: %q", role, len(got), len(want), got, want)
		}
		first, _, _ := strings.Cut(out, "\n")
		for _, w := range []string{"the " + role + " surface", "seat " + seatID, ": " + strconv.Itoa(len(want)-1) + " commands"} {
			if !strings.Contains(first, w) {
				t.Errorf("%s: the opening line %q does not say %q", role, first, w)
			}
		}
		// AND THE JOIN KEYS AGREE: every command the gates attribute to this role is a page here.
		have := map[string]bool{}
		for _, p := range pathsOf(pages) {
			have[p] = true
		}
		for _, k := range CommandPaths() {
			if typed, ok := strings.CutPrefix(k, role+" "); ok && !have[typed] {
				t.Errorf("%s: CommandPaths has %q and the manual has no page for %q", role, k, typed)
			}
		}
		if !have[manualName] {
			t.Errorf("%s: manual is not on its own surface", role)
		}
	}
}

// A SEAT'S MANUAL IS ITS OWN SURFACE AND NOBODY ELSE'S.
func TestManualNamesOnlyThisSeatsCommands(t *testing.T) {
	cases := []struct {
		seat          string
		has, hasNot   string
		foreignLeaves string // a role whose leaves must not appear unless this tree also has them
	}{
		{record.SampleSeatOf("blue"), "line-of-inquiry propose", "mint", "lens"},
		{record.SampleSeatOf("lens"), "mint", "line-of-inquiry propose", "blue"},
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

// EVERY PAGE IS THAT COMMAND'S OWN HELP, BYTE FOR BYTE, under a header naming it.
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
	pages := diagnostics.ManualPages(out, InvokedAs())
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

// THE SURVEY READS A MANUAL AS EVERY PAGE READ — the round trip from the writer's side, on real
// output. Without it, a seat that read its whole surface in one call would score every first use
// as run blind.
func TestTheSurveyReadsAManualAsEveryPageRead(t *testing.T) {
	seatID := record.SampleSeatOf("blue")
	out, _ := manualOf(t, seatID)
	bin := "/tmp/x/" + InvokedAs()
	lines := []any{
		map[string]any{"message": map[string]any{"content": []any{map[string]any{"type": "tool_use", "name": "Bash", "id": "a",
			"input": map[string]any{"command": `"` + bin + `" --seat-id ` + seatID + ` manual > /s/manual.txt`}}}}},
		map[string]any{"message": map[string]any{"content": []any{map[string]any{"type": "tool_result", "tool_use_id": "a", "content": out}}}},
		map[string]any{"message": map[string]any{"content": []any{map[string]any{"type": "tool_use", "name": "Bash", "id": "b",
			"input": map[string]any{"command": `"` + bin + `" line-of-inquiry propose --reason x`}}}}},
		map[string]any{"message": map[string]any{"content": []any{map[string]any{"type": "tool_result", "tool_use_id": "b", "content": "ok"}}}},
	}
	p := filepath.Join(t.TempDir(), "t.jsonl")
	var b []byte
	for _, l := range lines {
		j, _ := json.Marshal(l)
		b = append(append(b, j...), '\n')
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := diagnostics.ReadSurvey(p, InvokedAs(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.FirstUses) != 1 || s.FirstUses[0].Command != "line-of-inquiry propose" || s.FirstUses[0].Depth != diagnostics.DepthCommand {
		t.Errorf("first uses = %+v, want line-of-inquiry propose at depth command — the manual carried its page", s.FirstUses)
	}
	if got, want := len(s.HelpPages), len(surfaceOf(seatID)); got != want || s.ManualUnread != 0 {
		t.Errorf("survey saw %d pages (unread %d), the manual printed %d", got, s.ManualUnread, want)
	}
}
