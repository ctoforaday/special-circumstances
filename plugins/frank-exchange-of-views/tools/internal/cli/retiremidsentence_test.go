package cli

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A MARKER ON A MID-SENTENCE CLAUSE IS NEVER BARE, so the retire named nothing and said nothing.
// B9's G4: a corroboration marker sat on "not a hypothetical edge case, …" inside a sentence; blue
// edited the clause down to its markers and retired it twice, and both retires printed only
// "retired: …" while the marker stayed. The seat now names the marker, and the retire says what
// stayed when it names none.

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
	m := regexp.MustCompile(`the sky is blue<!--cite:(c-[0-9a-f]+)-->`).FindStringSubmatch(readReport(t, runDir))
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

func TestARetireOverAMidSentenceMarkerSaysItStayed(t *testing.T) {
	runDir, label := clauseCut(t)
	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "the sky is blue", "--reason", "refuted")
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); len(got) != 0 {
		t.Fatalf("a marker with prose before it in its sentence left without being named: %v", got)
	}
	if !strings.Contains(out, "anchor stayed: "+label) {
		t.Errorf("the retire took nothing and did not say the marker stayed:\n%s", out)
	}
}

func TestARetireNamingTheMarkerTakesItOut(t *testing.T) {
	runDir, label := clauseCut(t)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "the sky is blue",
		"--anchor", label, "--reason", "refuted"); err != nil {
		t.Fatalf("retire --anchor: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); !slices.Equal(got, []string{label}) {
		t.Fatalf("retire named %v, want [%s]", got, label)
	}
	if rep := readReport(t, runDir); strings.Contains(rep, label) {
		t.Errorf("the named marker is still in the report:\n%s", rep)
	}
}

func TestARetireRefusesAMarkerTheCutDidNotLeave(t *testing.T) {
	runDir, _ := clauseCut(t)
	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "the sky is blue",
		"--anchor", "c-00000000", "--reason", "refuted")
	if err == nil || !strings.Contains(err.Error(), "the report holds no") {
		t.Fatalf("a marker not in the report was accepted: %v", err)
	}
}

// A marker a later edit carried back into prose backs that text now, and leaves only with it —
// B9's edit 167 restored the clause around its markers.
func TestARetireRefusesAMarkerCarriedBackIntoProse(t *testing.T) {
	runDir, label := clauseCut(t)
	tok := "<!--cite:" + label + "-->"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "shows "+tok+" and", "--new", "shows the sky may be blue"+tok+" and", "--reason", "restored, qualified"); err != nil {
		t.Fatalf("restore prose around the marker: %v", err)
	}
	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "the sky is blue",
		"--anchor", label, "--reason", "refuted")
	if err == nil || !strings.Contains(err.Error(), "back into prose") {
		t.Fatalf("a marker re-attached to prose was taken out: %v", err)
	}
}
