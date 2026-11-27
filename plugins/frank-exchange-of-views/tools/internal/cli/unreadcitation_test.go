package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A VERIFICATION THAT READ NOTHING SAYS WHAT IT LEFT AND WHAT IT OWES: the failed check standing on
// the citation, worded as the work list words it — the citation stays unread, PASS is refused while
// it stands, and the lens that read nothing owes the gap about it; the gap minted,
// the verb names it instead.
func TestAnUnreadVerifySaysTheCitationStaysUnreadAndTheGapIsOwed(t *testing.T) {
	runDir, label := citedClause(t)
	out, err := verifyAs(t, runDir, label, "unreachable")
	if err != nil {
		t.Fatalf("verify --as unreachable: %v", err)
	}
	for _, want := range []string{"check of citation " + label, "read nothing of its source", "PASS is refused while it stands",
		lensSeat + " owes its gap about the anchor"} {
		if !strings.Contains(out, want) {
			t.Errorf("the verify does not say %q:\n%s", want, out)
		}
	}
	mintOut, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--about-kind", "anchor", "--about", label,
		"--class", "spec_underspecification", "--problem", "the source cannot be read", "--fix", "retire the citation or cite a readable source",
		"--check-kind", "document", "--check", "a lens reads the cited source", "--severity", "medium",
		"--likelihood", "medium", "--impact", "medium", "--complexity", "low")
	if err != nil {
		t.Fatalf("mint the gap about the unread citation: %v", err)
	}
	gap := gapID(mintOut)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
		t.Fatal(err)
	}
	other, err := run(t, "verify", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", label,
		"--as", "unreachable", "--confidence", "high", "--reason", "it would not load")
	if err != nil {
		t.Fatalf("a second lens's unreachable verify: %v", err)
	}
	if want := "gap " + gap + " carries it to blue"; !strings.Contains(other, want) {
		t.Errorf("the second unread verify does not say %q:\n%s", want, other)
	}
}

// --ANCHOR RETIRES ONE ANCHOR STANDING ON A SENTENCE, OF ANY KIND, AND NOTHING ELSE. Each refusal
// names what it was given and records nothing. The retire that lands changes no text, so it renders in
// no CHANGELOG section (lead call 11-26: anchor changes go to the work lists); `show changes` lists it
// as an anchor removal by retire. A finding anchor retires the same way.
func TestARetireTakesOneAnchorOffASentence(t *testing.T) {
	runDir := citedRun(t, "https://sky/1", "https://grass/1")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	sky := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	grass := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
	finding, _ := markerOfKind(t, runDir, "finding", "Fire is hot.")
	retire := func(anchors string) error {
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", anchors, "--reason", "unreadable")
		return err
	}
	refused := func(name string, err error, wants ...string) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: the retire was recorded", name)
		}
		for _, w := range wants {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s: the refusal does not say %q:\n%v", name, w, err)
			}
		}
	}
	retires := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
	refused("two citations", retire(sky+","+grass), "names 2", sky, grass)
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != retires {
		t.Errorf("%d retire(s) recorded by refused retires", n-retires)
	}

	if err := retire(sky); err != nil {
		t.Fatalf("retire --anchor %s: %v", sky, err)
	}
	refused("a retired citation", retire(sky), "the report holds no", sky)
	if err := retire(finding); err != nil {
		t.Fatalf("retire --anchor %s (a finding): %v", finding, err)
	}
	if md := readReport(t, runDir); !strings.Contains(md, "The sky is blue.") || strings.Contains(md, sky) || strings.Contains(md, finding) || !strings.Contains(md, "Fire is hot.") {
		t.Errorf("the retires did not take the two anchors off sentences that stay:\n%s", md)
	}
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	// A set with no edit has no CHANGELOG to write; either way it names no retire.
	changelog, err := os.ReadFile(filepath.Join(filepath.Dir(a.Report), report.FileChangelog))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, w := range []string{sky, finding, "The sky is blue", "Fire is hot", "retired from it"} {
		if strings.Contains(string(changelog), w) {
			t.Errorf("the changelog lists the anchor-only retire (%q); it changes no text:\n%s", w, changelog)
		}
	}
	changes, err := run(t, "show", "changes", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("show changes: %v", err)
	}
	for _, id := range []string{sky, finding} {
		if !strings.Contains(changes, id) {
			t.Errorf("show changes does not list the retire of %s:\n%s", id, changes)
		}
	}
}
