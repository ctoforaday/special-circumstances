package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A CLAIM REWRITTEN AWAY STILL OWNS ITS ANCHORS. Smoke m26: blue rewrote three anchored sentences
// into one new sentence carrying both finding markers (ev188). Retire credited only an edit whose
// replacement held nothing but anchors, so no retire could take those markers out — named, it was
// refused; split onto bare lines by a later edit, it took nothing and recorded the removal as
// asserted — and seven finding markers stood bare in the final report.

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

func TestARetireAfterARewriteSaysTheCarriedMarkerStayed(t *testing.T) {
	runDir, f1, _ := rewriteAway(t)
	// The marker stands before the full stop in the edit's old span, so the claim as a seat quotes
	// it is not a substring of that span's raw text.
	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", rwAlpha+".", "--reason", "gone")
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	ev := lastBody(t, runDir, &recordpb.Retire{})
	if len(ev.GetAnchors()) != 0 {
		t.Errorf("a marker carried onto standing prose left unnamed: %v", ev.GetAnchors())
	}
	if ev.GetRemovalBasis() != record.RemovalVerified {
		t.Errorf("removal_basis = %q for a claim a recorded edit's old span held", ev.GetRemovalBasis())
	}
	for _, want := range []string{"anchors out with it: none", "anchor stayed: " + f1} {
		if !strings.Contains(out, want) {
			t.Errorf("the retire's result does not say %q:\n%s", want, out)
		}
	}
}

func TestARetireAfterARewriteTakesTheMarkerItNames(t *testing.T) {
	runDir, f1, f2 := rewriteAway(t)
	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", rwAlpha+".", "--reason", "gone", "--anchor", f1)
	if err != nil {
		t.Fatalf("retire --anchor after a rewrite: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); !slices.Equal(got, []string{f1}) {
		t.Errorf("retire took %v, want [%s]", got, f1)
	}
	if !strings.Contains(out, "anchors out with it: "+f1) {
		t.Errorf("the result does not name the marker that left:\n%s", out)
	}
	rep := readReport(t, runDir)
	if strings.Contains(rep, anchor.Token(f1)) || !strings.Contains(rep, rwNew[:len(rwNew)-1]+anchor.Token(f2)) {
		t.Errorf("want %s out and %s still on the rewritten sentence:\n%s", f1, f2, rep)
	}
}

// splitToBareLines rewrites, then splits the carried markers onto lines of their own (m26 ev192).
func splitToBareLines(t *testing.T) (runDir, f1, f2 string) {
	t.Helper()
	runDir, f1, f2 = rewriteAway(t)
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", rwNew[:len(rwNew)-1]+anchor.Token(f1)+anchor.Token(f2)+".",
		"--new", rwNew+"\n\n"+anchor.Token(f1)+"\n\n"+anchor.Token(f2), "--reason", "split"); err != nil {
		t.Fatalf("split edit: %v", err)
	}
	return runDir, f1, f2
}

func TestARetireAfterARewriteTakesTheMarkerALaterEditLeftBare(t *testing.T) {
	runDir, f1, f2 := splitToBareLines(t)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", rwAlpha+".", "--reason", "gone"); err != nil {
		t.Fatalf("retire: %v", err)
	}
	ev := lastBody(t, runDir, &recordpb.Retire{})
	if !slices.Equal(ev.GetAnchors(), []string{f1}) {
		t.Errorf("retire took %v, want [%s] — the marker that stood on the claim and stands bare now", ev.GetAnchors(), f1)
	}
	if ev.GetRemovalBasis() != record.RemovalVerified {
		t.Errorf("removal_basis = %q", ev.GetRemovalBasis())
	}
	if rep := readReport(t, runDir); strings.Contains(rep, anchor.Token(f1)) || !strings.Contains(rep, anchor.Token(f2)) {
		t.Errorf("want %s out and %s standing:\n%s", f1, f2, rep)
	}
}

// A later edit whose replacement holds prose, but leaves THIS marker bare, has not put it back into
// prose: the check is of the marker, not of the edit.
func TestALaterEditThatLeftTheMarkerBareDoesNotHoldIt(t *testing.T) {
	runDir, f1, _ := splitToBareLines(t)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", rwAlpha+".", "--reason", "gone", "--anchor", f1); err != nil {
		t.Fatalf("retire --anchor of a marker a later edit left bare beside prose: %v", err)
	}
}

// An edit that did not hold the claim says nothing of it: a bare marker another cut left, waiting
// for its own retire, neither refuses this one nor leaves with it.
func TestAnotherClaimsBareMarkerDoesNotHoldARetire(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+rwAlpha+". Gamma stands alone.\n\nPlain prose here.\n")
	registerBlue(t, runDir)
	registerLensOnce(t, runDir)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", rwAlpha+".",
		"--reason", "unsupported", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatalf("finding: %v", err)
	}
	tok := anchor.Token(lastBody(t, runDir, &recordpb.Finding{}).GetId())
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", rwAlpha+tok+".", "--new", tok, "--reason", "cut"); err != nil {
		t.Fatalf("cut Alpha to its marker: %v", err)
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--quote", "Gamma stands alone.\n\nPlain prose here.", "--new", "Plain prose here.", "--reason", "cut"); err != nil {
		t.Fatalf("cut Gamma: %v", err)
	}
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "Gamma stands alone.", "--reason", "gone"); err != nil {
		t.Fatalf("a retire of an unanchored claim was held by another claim's bare marker: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Retire{}).GetAnchors(); len(got) != 0 {
		t.Errorf("the retire took %v, which never stood on its claim", got)
	}
}
