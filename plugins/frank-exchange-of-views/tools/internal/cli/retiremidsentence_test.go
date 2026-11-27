package cli

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A MARKER ON A MID-SENTENCE CLAUSE IS RETIRED BY NAME. B9's G4: a lens-placed citation marker sat on
// "not a hypothetical edge case, …" inside a sentence; blue edited the clause down to its markers and
// retired it twice, and both retires printed only "retired: …" while the marker stayed. A retire names
// its one anchor: it takes that marker off the sentence, which stays word for word, and refuses one the
// report does not hold.

// clauseCut cites a clause inside a sentence, edits the clause down to its marker, and returns the
// run and the marker's id. The rest of the sentence stands before the marker, so it is not bare.
func clauseCut(t *testing.T) (string, string) {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe data shows the sky is blue and the grass is green.\n\nClosing paragraph stays.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/m": []byte("<html>a source on the sky</html>")}})
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "the sky is blue", "--url", "https://sky/m", "--title", "Sky"); err != nil {
		t.Fatalf("cite the clause: %v", err)
	}
	m := regexp.MustCompile(`the sky is blue<!--cite:(C-[0-9a-f]{8})-->`).FindStringSubmatch(readReport(t, runDir))
	if m == nil {
		t.Fatalf("no marker after the clause:\n%s", readReport(t, runDir))
	}
	tok := "<!--cite:" + m[1] + "-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "the sky is blue"+tok, "--new", tok, "--reason", "the clause cannot stand"); err != nil {
		t.Fatalf("edit the clause down to its marker: %v", err)
	}
	return runDir, m[1]
}

func TestARetireNamingTheMarkerTakesItOut(t *testing.T) {
	runDir, label := clauseCut(t)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat,
		"--anchor", label, "--reason", "refuted"); err != nil {
		t.Fatalf("retire --anchor: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); !slices.Equal(got, []string{label}) {
		t.Fatalf("retire named %v, want [%s]", got, label)
	}
	if rep := readReport(t, runDir); strings.Contains(rep, label) || !strings.Contains(rep, "\n\nThe data shows and the grass is green.\n\n") {
		t.Errorf("want the named marker out and its sentence kept word for word:\n%s", rep)
	}
}

func TestARetireRefusesAMarkerTheCutDidNotLeave(t *testing.T) {
	runDir, _ := clauseCut(t)
	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat,
		"--anchor", "C-00000000", "--reason", "refuted")
	if err == nil || !strings.Contains(err.Error(), "the report holds no") {
		t.Fatalf("a marker not in the report was accepted: %v", err)
	}
}

// A MARKER A LATER EDIT CARRIED BACK INTO PROSE IS RETIRED THE SAME WAY (B9's edit 167 restored the
// clause around its markers). A retire takes an anchor off a sentence that stays, so the restored,
// qualified sentence keeps every word and loses only the named marker.
func TestARetireTakesAMarkerCarriedBackIntoProse(t *testing.T) {
	runDir, label := clauseCut(t)
	tok := "<!--cite:" + label + "-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "shows "+tok+" and", "--new", "shows the sky may be blue"+tok+" and", "--reason", "restored, qualified"); err != nil {
		t.Fatalf("restore prose around the marker: %v", err)
	}
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", label, "--reason", "refuted"); err != nil {
		t.Fatalf("a retire of a marker on prose was refused: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}); !slices.Equal(got.GetAnchors(), []string{label}) || got.GetClaim() != "The data shows the sky may be blue and the grass is green." {
		t.Errorf("the retire recorded anchors %v claim %q", got.GetAnchors(), got.GetClaim())
	}
	if rep := readReport(t, runDir); strings.Contains(rep, tok) || !strings.Contains(rep, "The data shows the sky may be blue and the grass is green.") {
		t.Errorf("want the marker out and the restored sentence kept word for word:\n%s", rep)
	}
}
