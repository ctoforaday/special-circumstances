package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A REWRITE CARRIES ITS ANCHORS, AND A RETIRE TAKES ONE BY NAME. Smoke m26: blue rewrote three anchored
// sentences into one new sentence carrying both finding markers (ev188), then split the markers onto
// bare lines (ev192), and seven finding markers stood bare in the final report. A rewrite carries
// every anchor onto prose; a retire takes one of them off the sentence, which stays word for word; and
// an edit that would leave the carried markers on no words is refused, so none stands bare.

const (
	rwAlpha = "Alpha holds for every input"
	rwBeta  = "Beta holds by the same argument"
	rwLimit = "The search covered one query"
	rwNew   = "Whether more holds is not settled here."
)

// rewriteAway files a finding on each of two sentences, then rewrites the paragraph's three
// sentences into one new sentence the two markers are carried onto. It returns the run and the
// two finding ids.
func rewriteAway(t *testing.T) (runDir, f1, f2 string) {
	t.Helper()
	runDir = newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+rwAlpha+". "+rwBeta+". "+rwLimit+".\n\nPlain prose here.\n")
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)
	find := func(key, quote string) string {
		t.Helper()
		if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
			"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatalf("finding %s: %v", key, err)
		}
		return lastBody(t, runDir, &recordpb.Finding{}).GetId()
	}
	f1, f2 = find("F1", rwAlpha+"."), find("F2", rwBeta+".")
	rep := readReport(t, runDir)
	span := rep[strings.Index(rep, rwAlpha):strings.Index(rep, "\n\nPlain prose")]
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", span, "--new", rwNew+anchor.Token(f1)+anchor.Token(f2), "--reason", "rewrite"); err != nil {
		t.Fatalf("rewrite edit: %v", err)
	}
	return runDir, f1, f2
}

func TestARetireAfterARewriteTakesTheMarkerItNames(t *testing.T) {
	runDir, f1, f2 := rewriteAway(t)
	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "gone", "--anchor", f1)
	if err != nil {
		t.Fatalf("retire --anchor after a rewrite: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); !slices.Equal(got, []string{f1}) {
		t.Errorf("retire took %v, want [%s]", got, f1)
	}
	if !strings.Contains(out, "retired finding "+f1) {
		t.Errorf("the result does not name the marker that left:\n%s", out)
	}
	rep := readReport(t, runDir)
	if strings.Contains(rep, anchor.Token(f1)) || !strings.Contains(rep, rwNew[:len(rwNew)-1]+anchor.Token(f2)) {
		t.Errorf("want %s out and %s still on the rewritten sentence:\n%s", f1, f2, rep)
	}
}

// AN EDIT SPLITTING THE CARRIED MARKERS ONTO LINES OF THEIR OWN IS REFUSED (m26 ev192), naming both,
// and nothing is recorded: a marker on no words marks no sentence.
func TestASplitLeavingCarriedMarkersBareIsRefused(t *testing.T) {
	runDir, f1, f2 := rewriteAway(t)
	before, edits := readReport(t, runDir), countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", rwNew[:len(rwNew)-1]+anchor.Token(f1)+anchor.Token(f2)+".",
		"--new", rwNew+"\n\n"+anchor.Token(f1)+"\n\n"+anchor.Token(f2), "--reason", "split")
	if err == nil {
		t.Fatalf("a split leaving %s and %s on lines of their own landed:\n%s", f1, f2, readReport(t, runDir))
	}
	for _, w := range []string{anchor.Token(f1), anchor.Token(f2), "standing bare", "keep words on", "cut it whole"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
	if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != edits {
		t.Error("the refused split changed the report or the record")
	}
}
