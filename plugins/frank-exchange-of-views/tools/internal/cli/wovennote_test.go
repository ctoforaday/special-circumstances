package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A FOOTNOTE TYPED UNDER A LABEL THE WEAVE WRITES IS REFUSED AT THE WRITE, through the real verbs.
//
// Assembly numbers a citation's note [^1], [^2] … and a proof's [^P1], [^P2] …, and writes them
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
		"`[^1]`, `[^2]` … for citations and `[^P1]`, `[^P2]` … for proofs",
		"in report.md, from the citations and proofs anchored in the text",
		"a reader is shown one of the two",
		"NOTHING IS RECORDED",
		"HOW CONTENT REACHES IT: Blue's `cite` places a citation on the sentence its source backs, and blue's `prove` a proof on the sentence it establishes.",
		"WHAT TO DO: take every reference of that shape, and the `[^…]:` line that defines it, out of the report's text",
		"An aside that is neither a source nor a computation is prose, or a parenthesis, where it applies",
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
}

func TestIngestRefusesAFootnoteUnderALabelTheWeaveWrites(t *testing.T) {
	const head = "# On 91\n\n## TL;DR\n\n"
	for _, tc := range []struct{ name, body, label string }{
		{"a citation's label, reference and definition", "91 is 7 x 13.[^1]\n\n[^1]: A note about seven.\n", "[^1]"},
		{"a proof's label, the definition alone", "91 is 7 x 13.\n\n[^P2]: A note about thirteen.\n", "[^P2]"},
		{"a reference with no definition", "91 is 7 x 13[^12].\n", "[^12]"},
		{"a proof's label in lower case", "91 is 7 x 13[^p1].\n\n[^p1]: A note.\n", "[^p1]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			path := filepath.Join(runDir, "blue", "report.md")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(head+tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
				t.Fatalf("register: %v", err)
			}
			out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
			refusesTheNote(t, out, err, tc.label)
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BASE_INGEST); n != 0 {
				t.Errorf("the refused ingest recorded %d base(s)", n)
			}
			if _, err := os.Stat(path); err != nil {
				t.Errorf("the refused ingest did not leave the file for its author to fix: %v", err)
			}
		})
	}
}

func TestEditRefusesAFootnoteUnderALabelTheWeaveWrites(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.[^scope] It is odd. The draft read [^P3 sic.\n\n[^scope]: Of the natural numbers.\n")
	registerBlue(t, runDir)

	// The label whole, in --new.
	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "It is odd.", "--new", "It is odd.[^2]", "--reason", "note it")
	refusesTheNote(t, out, err, "[^2]")
	// And a label the edit COMPLETES: no whole label is in --new, so only the planned document shows it.
	out, err = run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E2",
		"--quote", "P3 sic", "--new", "P3]", "--reason", "renumber")
	refusesTheNote(t, out, err, "[^P3]")
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
// a proof placed beside it, and reaches the reader once in each tier with its own definition.
func TestAFootnoteOfTheReportsOwnShipsBesideTheWovenNotes(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.[^scope] Trial division finds 7. Thirteen is prime.\n\n[^scope]: Of the natural numbers.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://tables/1": []byte("<html>division tables</html>")}})
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", "Trial division finds 7.",
		"--url", "https://tables/1", "--title", "Division tables", "--reason", "table two lists the quotient"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if out, err := run(t, "prove", "--run", runDir, "--seat-id", blueSeat, "--quote", "Thirteen is prime.",
		"--script", script(t, runDir, "prime.js", "console.log(13)"), "--reason", "thirteen has no divisor below four"); err != nil {
		t.Fatalf("prove: %v\n%s", err, out)
	}
	open := runtest.Open(t, runDir)
	docs, err := report.AssembleAll(open)
	if err != nil {
		t.Fatal(err)
	}
	var shipped string
	for _, d := range docs {
		if d.File == report.FileReport {
			shipped = d.Body
		}
	}
	for _, want := range []string{"91 is 7 x 13.[^scope]", "\n[^scope]: Of the natural numbers.\n", "finds 7[^1].", "\n[^1]: Division tables.", "is prime[^P1].", "\n[^P1]: thirteen has no divisor below four"} {
		if n := strings.Count(shipped, want); n != 1 {
			t.Errorf("report.md carries %q %d time(s), want once:\n%s", want, n, shipped)
		}
	}
	html := report.RenderSite(report.Title(open), docs, record.Family{})
	for id, text := range map[string]string{"fn-scope": "Of the natural numbers.", "fn-1": "Division tables.", "fn-P1": "thirteen has no divisor below four"} {
		li := `<li id="` + id + `">`
		if n := strings.Count(html, li); n != 1 || !strings.Contains(html, li+text) {
			t.Errorf("report.html holds %d entr(ies) %s, want one that reads %q", n, li, text)
		}
	}
	// What the weave wrote is what the write refuses: the two are one set of labels.
	if label, _, found := report.WovenNoteIn(shipped); !found || label != "[^1]" {
		t.Errorf("the assembled report's first woven label reads %q (found %v), want [^1]", label, found)
	}
	if label, _, found := report.WovenNoteIn(readReport(t, runDir)); found {
		t.Errorf("the report's text carries %s, a label the weave writes", label)
	}
}
