package cli

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A REPORT'S OWN "## Computations" AND "## Motions" ARE THE REPORT'S, through the real verbs.
//
// The tool composes sections of those names in evidence.md and judgments.md, never in report.md,
// and both are what a subject calls a section of its own: a report on an algorithm has computations, one on
// an assembly has motions. Text under either meets no section of the tool's, so it ingests, edits,
// and ships whole and once among the report's own sections, while the tool's two sections are
// composed from the record exactly as they are without it.
func TestAReportsOwnComputationsAndMotionsShipAsTheReports(t *testing.T) {
	ownComputations := []string{"Lemma one: seven divides ninety-one.", "Lemma two: thirteen divides ninety-one.", "- The quotient of the first division is thirteen."}
	ownMotions := []string{"The assembly moved to adopt the factorisation.", "| Motion | Carried |", "| --- | --- |", "| Adopt | ✓ |"}

	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Computations\n\n"+ownComputations[0]+"\n\n"+ownComputations[2]+"\n\n## Resolutions\n\n"+strings.Join(ownMotions, "\n")+"\n")
	registerBlue(t, runDir)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)

	// An edit under the heading, and an edit that writes the other heading.
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", ownComputations[0], "--new", ownComputations[0]+"\n\n"+ownComputations[1], "--reason", "the second factor"); err != nil {
		t.Fatalf("an edit under the report's own `## Computations` was refused: %v\n%s", err, out)
	}
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E2",
		"--quote", "Resolutions", "--new", "Motions", "--reason", "what the assembly calls them"); err != nil {
		t.Fatalf("an edit writing the report's own `## Motions` was refused: %v\n%s", err, out)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 2 {
		t.Fatalf("%d edit(s) on the record, want both", n)
	}

	// The tool's own sections of those names, with something under each.
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "91 is 7 x 13.",
		"--script", script(t, runDir, "factor.js", "console.log(7*13)"), "--reason", "the product of the two factors is computed"); err != nil {
		t.Fatalf("compute: %v\n%s", err, out)
	}
	if out, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", lensSeat,
		"--class", "safety", "--relief", "stop and review", "--reason", "the report instructs a hazardous procedure"); err != nil {
		t.Fatalf("petition file: %v\n%s", err, out)
	}

	open := runtest.Open(t, runDir)
	docs, err := shippedDocs(open)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]string{}
	for _, d := range docs {
		body[d.File] = d.Body
	}
	shipped := body[report.FileReport]
	at := strings.Index(shipped, "## TL;DR")
	if at < 0 || strings.Contains(shipped, "## Additional findings") {
		t.Fatalf("report.md has no TL;DR, or a heading of the tool's above the report's own sections:\n%s", shipped)
	}
	for _, h := range []string{"## Computations", "## Motions"} {
		if n := strings.Count(shipped, h+"\n"); n != 1 || strings.Index(shipped, h+"\n") < at {
			t.Errorf("report.md carries %q %d time(s), want once, in the body:\n%s", h, n, shipped)
		}
	}
	// CHANGELOG.md quotes each edit's spans as history, the two edits above among them; that is a
	// record of the edit, not a second section carrying the report's lines.
	for _, ln := range slices.Concat(ownComputations, ownMotions) {
		n := 0
		for file, b := range body {
			if file == report.FileChangelog {
				continue
			}
			c := strings.Count(b, ln)
			n += c
			if c > 0 && file != report.FileReport {
				t.Errorf("%s carries the report's own line %q", file, ln)
			}
		}
		if n != 1 || strings.Index(shipped, ln) < at {
			t.Errorf("the set carries the report's own line %q %d time(s), want once, in report.md's body", ln, n)
		}
	}
	// The tool's sections: one of each, in its own document, holding what the record holds.
	if b := body[report.FileEvidence]; strings.Count(b, "## Computations\n") != 1 || !strings.Contains(b, "console.log(7*13)") || !strings.Contains(b, "the product of the two factors is computed") {
		t.Errorf("evidence.md does not carry the tool's Computations section with the computation on the record:\n%s", b)
	}
	if b := body[report.FileJudgments]; strings.Count(b, "## Motions\n") != 1 || !strings.Contains(b, "stop and review") {
		t.Errorf("judgments.md does not carry the tool's Motions section with the motion on the record:\n%s", b)
	}

	// The layout says both things of each name: the tool's section in its document, and whose a
	// section of that name in the report's text is.
	view, err := report.Layout(open)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range view.Sections {
		switch s.Heading {
		case "Computations", "Motions":
			if !s.OwnInReport || !s.Ships || s.Composer != report.Tool || s.File == report.FileReport || strings.Contains(s.Text, "Lemma") || strings.Contains(s.Text, "assembly moved") {
				t.Errorf("the layout's %s row: %+v", s.Heading, s)
			}
		case "The report's own sections":
			if !slices.Equal(s.AuthoredHeadings, []string{"TL;DR", "Computations", "Motions"}) {
				t.Errorf("the body holds %v, want the report's own TL;DR, Computations and Motions", s.AuthoredHeadings)
			}
		default:
			if s.OwnInReport {
				t.Errorf("the layout says a `## %s` in the report's text is the report's own", s.Heading)
			}
		}
	}
	out, err := run(t, "show", "report", "--assembled", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "`ingest` and `edit` refuse one. The exception is a section of another document whose entry says that a section of that name in the report's text is the report's own.") {
		t.Errorf("the assembled layout says the report's text carries no section the tool composes, and does not say which entries are the exception:\n%s", out)
	}
	if n := strings.Count(out, report.OwnInTheReport); n != 2 {
		t.Errorf("the assembled layout says %q %d time(s), want once under each of Computations and Motions:\n%s", report.OwnInTheReport, n, out)
	}

	// ONE PAGE, ONE ELEMENT PER ID. report.html holds every document as a tab, and a fragment
	// resolves to the first element of that id on the page: the report's own "Computations" and
	// evidence.md's are two ids.
	set, err := report.AssembleShipped(open)
	if err != nil {
		t.Fatal(err)
	}
	html := report.RenderSite(set, record.Family{})
	seen := map[string]bool{}
	for _, m := range regexp.MustCompile(` id="([^"]+)"`).FindAllStringSubmatch(html, -1) {
		if seen[m[1]] {
			t.Errorf("report.html holds two elements with id %q", m[1])
		}
		seen[m[1]] = true
	}
	for _, id := range []string{"computations", "computations-2", "motions", "motions-2"} {
		if !seen[id] {
			t.Errorf("report.html has no element with id %q", id)
		}
	}
}

// THE SAME NAMES ARE REFUSED WHERE THE ROW SAYS SO: taking the field off a row refuses the
// heading, and the refusal is the one every composed section gets.
func TestTheBoardIsStillRefusedInTheReportsText(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Standing\n\nNothing is open.\n")
	registerBlue(t, runDir)
	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "Standing", "--new", "The board", "--reason", "name it")
	refusesTheSection(t, out, err, `"## The board"`, "docket.md", "No act writes it as text")
}
