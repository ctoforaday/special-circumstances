package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"google.golang.org/protobuf/proto"
)

// A FOOTNOTE TYPED UNDER A LABEL THE WEAVE WRITES IS REFUSED AT THE WRITE, through the real verbs.
//
// Assembly numbers a citation's note [^1], [^2] … and a computation's [^K1], [^K2] …, and writes them
// into the report's text as it stands. A hand note under one of those labels shipped beside the
// tool's on a copy of universe-m20: report.md held two definitions of [^1] and report.html one
// entry for it, so the reader got one note and nothing said the other existed. The author reads
// the refusal at ingest and at edit; it names the label, why that shape is the tool's, and the
// acts that put a source or a computation on a sentence.

func refusesTheNote(t *testing.T, out string, err error, label string) {
	t.Helper()
	if err == nil {
		t.Fatalf("report text carrying the footnote label %s was recorded:\n%s", label, out)
	}
	msg := err.Error() + "\n" + out
	for _, w := range []string{
		`the footnote label "` + label + `"`,
		"`[^1]`, `[^2]` … for citations and `[^K1]`, `[^K2]` … for computations",
		"in report.md, from the citations and computations anchored in the text",
		"a reader is shown one of the two",
		"NOTHING IS RECORDED",
		"HOW CONTENT REACHES IT: Blue's `cite` places a citation on the sentence its source backs, and blue's `compute` a computation on the sentence it establishes. No note is typed by hand under a number, or under K and a number: a footnote of the report's own — an aside that is neither a source nor a computation — takes a label that is a word.",
		"WHAT TO DO: take every reference of that shape this write adds, and the `[^…]:` line that defines it, out of the text. A source or a computation goes by the way above.",
		"Where such a label is the subject's own text — a character class in a regular expression — it goes in a code span or a fenced block, and is no footnote there",
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
}

func TestIngestRefusesAFootnoteUnderALabelTheWeaveWrites(t *testing.T) {
	const head = "## The factors\n\n"
	for _, tc := range []struct{ name, body, label string }{
		{"a citation's label, reference and definition", "91 is 7 x 13.[^1]\n\n[^1]: A note about seven.\n", "[^1]"},
		{"a computation's label, the definition alone", "91 is 7 x 13.\n\n[^K2]: A note about thirteen.\n", "[^K2]"},
		{"a reference with no definition", "91 is 7 x 13[^12].\n", "[^12]"},
		{"a computation's label in lower case", "91 is 7 x 13[^k1].\n\n[^k1]: A note.\n", "[^k1]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := laneRun(t, 1)
			writePart(t, runDir, "blue-lane-1", head+tc.body)
			out, err := ingestPart(t, runDir, "blue-lane-1")
			refusesTheNote(t, out, err, tc.label)
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPORT_PART); n != 0 {
				t.Errorf("the refused ingest recorded %d part(s)", n)
			}
			if _, err := os.Stat(partFile(runDir, "blue-lane-1")); err != nil {
				t.Errorf("the refused ingest did not leave the file for its author to fix: %v", err)
			}
		})
	}
}

func TestEditRefusesAFootnoteUnderALabelTheWeaveWrites(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.[^scope] It is odd. The draft read [^K3 sic.\n\n[^scope]: Of the natural numbers.\n")
	registerBlue(t, runDir)

	// The label whole, in --new.
	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "It is odd.", "--new", "It is odd.[^2]", "--reason", "note it")
	refusesTheNote(t, out, err, "[^2]")
	// And a label the edit COMPLETES: no whole label is in --new, so only the planned document shows it.
	out, err = run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E2",
		"--quote", "K3 sic", "--new", "K3]", "--reason", "renumber")
	refusesTheNote(t, out, err, "[^K3]")
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Errorf("the refused edits recorded %d edit(s)", n)
	}
	// The same characters in code are not a footnote.
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E3",
		"--quote", "It is odd.", "--new", "It is odd. The class `[^1]` matches neither `^` nor `1`.\n\n```\n[^1]: not a note\n```", "--reason", "show the class"); err != nil {
		t.Fatalf("a label of the weave's shape inside code was refused: %v\n%s", err, out)
	}
}

// A FOOTNOTE OF THE REPORT'S OWN, UNDER A LABEL OF ANOTHER SHAPE, SHIPS AND RESOLVES. Only the
// weave's labels can meet the weave's notes; this one is ingested, stands through a citation and
// a computation placed beside it, and reaches the reader once in each tier with its own definition.
func TestAFootnoteOfTheReportsOwnShipsBesideTheWovenNotes(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.[^scope] Trial division finds 7. Thirteen is prime.\n\n[^scope]: Of the natural numbers.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://tables/1": []byte("<html>division tables</html>")}})
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", "Trial division finds 7.",
		"--url", "https://tables/1", "--title", "Division tables", "--reason", "table two lists the quotient"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Thirteen is prime.",
		"--dir", script(t, runDir, "prime.js", "console.log(13)"), "--reason", "thirteen has no divisor below four", "--", "node", "prime.js"); err != nil {
		t.Fatalf("compute: %v\n%s", err, out)
	}
	open := runtest.Open(t, runDir)
	docs, err := shippedDocs(open)
	if err != nil {
		t.Fatal(err)
	}
	var shipped string
	for _, d := range docs {
		if d.File == report.FileReport {
			shipped = d.Body
		}
	}
	for _, want := range []string{"91 is 7 x 13.[^scope]", "\n[^scope]: Of the natural numbers.\n", "finds 7[^1].", "\n[^1]: Division tables.", "is prime[^K1].", "\n[^K1]: thirteen has no divisor below four"} {
		if n := strings.Count(shipped, want); n != 1 {
			t.Errorf("report.md carries %q %d time(s), want once:\n%s", want, n, shipped)
		}
	}
	set, err := report.AssembleShipped(open)
	if err != nil {
		t.Fatal(err)
	}
	html := report.RenderSite(set, record.Family{})
	for id, text := range map[string]string{"fn-scope": "Of the natural numbers.", "fn-1": "Division tables.", "fn-K1": "thirteen has no divisor below four"} {
		li := `<li id="` + id + `">`
		if n := strings.Count(html, li); n != 1 || !strings.Contains(html, li+text) {
			t.Errorf("report.html holds %d entr(ies) %s, want one that reads %q", n, li, text)
		}
	}
	// What the weave wrote is what the write refuses: the two are one set of labels.
	if label, _, found := report.WovenNoteIn(report.ParseText(""), report.ParseText(shipped)); !found || label != "[^1]" {
		t.Errorf("the assembled report's first woven label reads %q (found %v), want [^1]", label, found)
	}
	if label, _, found := report.WovenNoteIn(report.ParseText(""), report.ParseText(readReport(t, runDir))); found {
		t.Errorf("the report's text carries %s, a label the weave writes", label)
	}
}

// AN EDIT IS REFUSED FOR WHAT IT ADDS, NOT FOR WHAT THE RECORD HOLDS. A record ingested before the
// refusals existed, or migrated, holds a heading and labels no live write makes. Asked of the whole
// planned document, the refusal fired on every edit of such a report — the edit that takes one
// label out included, because the next still stood, and a footnote reference is not text a quote
// can name on its own — so nothing could be recorded on that run again. The report as it stood is
// the allowance: an edit elsewhere is recorded, one that takes the claims out is recorded, and one
// that adds a claim, a second copy of one the record holds included, is refused.
func TestAnEditIsRefusedForWhatItAddsNotForWhatTheRecordHolds(t *testing.T) {
	runDir := newRun(t)
	const held = "# On 91\n\n## TL;DR\n\n91 is 7 x 13.[^1] It is odd.[^K2] Thirteen is prime.\n\n## The board\n\nSeven divides it. Nothing else does[^3].\n\n## The debate\n\nNone.\n\n[^1]: A hand note about seven.\n"
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}, &recordpb.ReportPart{Text: proto.String(held)}); err != nil {
		t.Fatal(err)
	}
	registerBlue(t, runDir)
	edit := func(key, quote, text string) (string, error) {
		return run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key, "--quote", quote, "--new", text, "--reason", "r")
	}

	// An edit that touches none of it.
	if out, err := edit("E1", "Thirteen is prime.", "Thirteen is a prime."); err != nil {
		t.Fatalf("an edit that adds no heading and no label was refused for what the record holds: %v\n%s", err, out)
	}
	// One that adds: a label of a new number, a second reference under a label the record holds,
	// a second section under a heading the record holds, and a heading it does not hold.
	// The new label stands on a line that already holds two, after them.
	out, err := edit("E2", "Thirteen is a prime.", "Thirteen is a prime.[^7]")
	refusesTheNote(t, out, err, "[^7]")
	out, err = edit("E3", "Seven divides it.", "Seven divides it.[^1]")
	refusesTheNote(t, out, err, "[^1]")
	out, err = edit("E4", "None.", "None.\n\n## The board\n\nAgain.")
	refusesTheSection(t, out, err, `"## The board"`, "in docket.md")
	out, err = edit("E5", "None.", "None.\n\n## Risk matrix\n\nNone.")
	refusesTheSection(t, out, err, `"## Risk matrix"`)
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 1 {
		t.Errorf("the record holds %d edit(s), want the one that added nothing", n)
	}
	// One that moves a label, and ones that take the claims out, each with the others standing.
	for _, e := range [][3]string{
		{"E6", "It is odd.[^K2] Thirteen", "It is odd. Thirteen[^p2]"},
		{"E7", "91 is 7 x 13.[^1] It is odd. Thirteen[^p2] is", "91 is 7 x 13. It is odd. Thirteen is"},
		{"E8", "## The board", "## THE  BOARD"}, // the same claim, spelled another way
		{"E10", "## THE  BOARD", "## What divides it"},
		{"E9", "## The debate\n\nNone.", "Nothing more."},
	} {
		if out, err := edit(e[0], e[1], e[2]); err != nil {
			t.Fatalf("edit %s, which adds no heading and no label, was refused: %v\n%s", e[0], err, out)
		}
		if e[0] == "E8" {
			// The spelling is another and the section is the one the record held: a second is still one added.
			out, err := edit("E8b", "Nothing else does", "Nothing else does.\n\n## The board\n\nAgain")
			refusesTheSection(t, out, err, `"## The board"`)
		}
	}
	got := readReport(t, runDir)
	if want := "# On 91\n\n## TL;DR\n\n91 is 7 x 13. It is odd. Thirteen is a prime.\n\n## What divides it\n\nSeven divides it. Nothing else does[^3].\n\nNothing more.\n\n[^1]: A hand note about seven.\n"; got != want {
		t.Errorf("the report reads\n%q\nwant\n%q", got, want)
	}
}
