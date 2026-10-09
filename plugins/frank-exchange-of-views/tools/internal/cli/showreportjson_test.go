package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// reportDoc is what `show report --json` prints: the projection, and the selection block a selector adds.
type reportDoc struct {
	record.ReportJSON
	Selection *struct {
		Criterion string `json:"criterion"`
		Matched   int    `json:"matched"`
		Of        int    `json:"of"`
		Complete  bool   `json:"complete"`
	} `json:"selection"`
}

// reportJSONOf is `show report --json` with any further flags, parsed.
func reportJSONOf(t *testing.T, runDir string, extra ...string) reportDoc {
	t.Helper()
	out, err := run(t, append([]string{"show", "report", "--run", runDir, "--seat-id", lensSeat, "--json"}, extra...)...)
	if err != nil {
		t.Fatalf("show report --json %v: %v", extra, err)
	}
	var d reportDoc
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("show report --json %v is not JSON (%v):\n%s", extra, err, out)
	}
	return d
}

// `show report --json` IS THE REPORT'S OWN LINES AT ITS HEAD, UNDER THE SELECTORS THE MARKDOWN TAKES.
//
// The shared --json flag is on every projection's page and `report` refused it: 13 refusals on
// universe m18 and 7 on m17, and one seat read the refusal's envelope as the report. Each form is
// held to the markdown read of the same record, so the JSON cannot become a second report.
func TestShowReportJSONIsTheSameLinesAtHead(t *testing.T) {
	const base = "# Costs\n\nQ1 was hard. Costs rose sharply in the first quarter, then fell.\n\nVolume grew.\n\n# Other\n\nNothing here.\n"
	runDir := newRun(t)
	writeReport(t, runDir, base)
	id, err := mintQuote(t, runDir, "k1", "Costs rose sharply")
	if err != nil {
		t.Fatal(err)
	}
	bare := readReport(t, runDir)

	t.Run("whole: the lines join back to the markdown read", func(t *testing.T) {
		d := reportJSONOf(t, runDir)
		var text []string
		for i, ln := range d.Lines {
			if ln.Line != i+1 {
				t.Fatalf("line %d is numbered %d", i+1, ln.Line)
			}
			text = append(text, ln.Text)
		}
		if got := strings.Join(text, "\n"); got != bare {
			t.Errorf("the lines do not join back to the report:\n json %q\n bare %q", got, bare)
		}
		if d.Selection != nil {
			t.Errorf("a read with no selector carries a selection block: %+v", d.Selection)
		}
	})
	t.Run("head: an edit moves it, with the text", func(t *testing.T) {
		before := reportJSONOf(t, runDir)
		if before.Head == 0 {
			t.Fatal("an ingested report reads head 0")
		}
		registerBlue(t, runDir)
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
			"--quote", "Nothing here.", "--new", "Something here.", "--reason", "a head to move"); err != nil {
			t.Fatalf("edit: %v", err)
		}
		after := reportJSONOf(t, runDir)
		if after.Head <= before.Head {
			t.Errorf("an edit left the head at %d (was %d)", after.Head, before.Head)
		}
		if n := len(after.Lines); n < 2 || after.Lines[n-2].Text != "Something here." {
			t.Errorf("the text at the new head does not hold the edit: %+v", after.Lines)
		}
		bare = readReport(t, runDir)
	})
	t.Run("--quote: the hit lines with their heading, and what was kept of what there was", func(t *testing.T) {
		d := reportJSONOf(t, runDir, "--quote", "Volume grew.")
		if len(d.Lines) != 1 || d.Lines[0].Text != "Volume grew." || d.Lines[0].Heading != "# Costs" || d.Lines[0].Line != 5 {
			t.Fatalf("want line 5 under # Costs, got %+v", d.Lines)
		}
		if d.Selection == nil || d.Selection.Matched != 1 || d.Selection.Of != len(strings.Split(bare, "\n")) || !d.Selection.Complete {
			t.Errorf("the selection block does not say 1 of the report's lines: %+v", d.Selection)
		}
	})
	t.Run("--match that hits nothing: an empty list beside the report's size, never an absent key", func(t *testing.T) {
		out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--json", "--match", "zebra")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, `"lines": []`) {
			t.Errorf("a selection that kept nothing does not print `lines: []`:\n%s", out)
		}
		if d := reportJSONOf(t, runDir, "--match", "zebra"); d.Selection == nil || d.Selection.Matched != 0 || d.Selection.Of == 0 {
			t.Errorf("no match reads as an empty report: %+v", d.Selection)
		}
	})
	t.Run("--anchor: the window the markdown read shows, the anchored line in it", func(t *testing.T) {
		d := reportJSONOf(t, runDir, "--anchor", id, "--window", "1")
		md, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--anchor", id, "--window", "1")
		if err != nil {
			t.Fatal(err)
		}
		holds := false
		for _, ln := range d.Lines {
			if !strings.Contains(md, ln.Text) {
				t.Errorf("the JSON window holds a line the markdown window does not: %q", ln.Text)
			}
			holds = holds || strings.Contains(ln.Text, anchor.Token(id))
		}
		if !holds || len(d.Lines) >= len(strings.Split(bare, "\n")) {
			t.Errorf("the window is not a window on the anchor: %+v", d.Lines)
		}
	})
	t.Run("an anchor that is not in the report is refused in both forms", func(t *testing.T) {
		if _, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--json", "--anchor", "G-00000000"); err == nil {
			t.Error("--json --anchor on an anchor the report does not hold answered")
		}
	})
}
