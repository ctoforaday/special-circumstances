package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// closeQuoteGap closes gap as repaired with the verification triple, the lens's own close.
func closeQuoteGap(t *testing.T, runDir, gap string) {
	t.Helper()
	if _, err := run(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes"); err != nil {
		t.Fatalf("close %s: %v", gap, err)
	}
}

// RETIRE TAKES ONE ANCHOR OF ANY KIND AND KEEPS ITS SENTENCE (P12b, ruling 3). The verb is
// `retire --anchor <one id> --reason … [--answers …]`: the tool stamps the sentence the anchor stood on
// as the retire's claim, the anchor leaves the report, and the sentence stays word for word. The claim
// retire is gone: --quote and --new are no flags of the verb, and --anchor is required.
func TestRetireIsAnchorOnlyForEveryKind(t *testing.T) {
	const sentence = "Water is wet."
	for _, kind := range anchor.Kinds() {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			tok := anchor.Token(id)
			if !strings.Contains(readReport(t, runDir), "Water is wet"+tok) {
				t.Fatalf("no %s anchor on the bullet:\n%s", kind, readReport(t, runDir))
			}
			if kind == "gap" {
				// An open gap's own marker is refused (TestAnOpenGapsMarkerIsNotRetired); closed, it may go.
				closeQuoteGap(t, runDir, id)
			}
			out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "it does not hold", "--anchor", id)
			if err != nil {
				t.Fatalf("retire --anchor %s (%s): %v", id, kind, err)
			}
			ev := lastBody(t, runDir, &recordpb.Retire{})
			if !slices.Equal(ev.GetAnchors(), []string{id}) || ev.GetClaim() != sentence || ev.GetReason() != "it does not hold" {
				t.Errorf("the retire recorded anchors %v claim %q reason %q, want [%s], %q, the reason given", ev.GetAnchors(), ev.GetClaim(), ev.GetReason(), id, sentence)
			}
			if !strings.Contains(out, id) || !strings.Contains(out, "its sentence stays in the report word for word") {
				t.Errorf("the retire's result does not name %s and the sentence kept:\n%s", id, out)
			}
			if rep := readReport(t, runDir); strings.Contains(rep, tok) || !strings.Contains(rep, "- "+sentence+"\n") {
				t.Errorf("after the retire the report should hold %q word for word and not %s:\n%s", sentence, tok, rep)
			}
		})
	}

	// THE CLAIM RETIRE IS GONE: no --quote, no --new, and no retire without --anchor.
	runDir := citedRun(t, "https://water/1")
	before := readReport(t, runDir)
	for _, args := range [][]string{
		{"--quote", sentence, "--reason", "refuted"},
		{"--anchor", "C-00000001", "--new", "a successor", "--reason", "refuted"},
		{"--reason", "refuted"},
	} {
		_, err := run(t, append([]string{"retire", "--run", runDir, "--seat-id", blueSeat}, args...)...)
		if err == nil {
			t.Errorf("retire %v was recorded", args)
			continue
		}
		if !strings.Contains(err.Error(), "unknown flag") && !strings.Contains(err.Error(), "anchor") {
			t.Errorf("retire %v was refused for another reason: %v", args, err)
		}
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
		t.Errorf("%d retire(s) recorded by refused calls", n)
	}
	if after := readReport(t, runDir); after != before {
		t.Errorf("a refused retire changed the report:\n%s", after)
	}
}

// F14: A RETIRE OF AN OPEN GAP'S OWN MARKER IS REFUSED, NAMING THE GAP (ruled by gblock 2026-11-20).
// The retire keeps the sentence, so the gap would stand on no marker while it is owed; a cut of the
// sentence is the gap's raiser's turn instead. Nothing is recorded, and the report is unchanged. Once the
// gap is closed, its marker may be retired off the sentence.
func TestAnOpenGapsMarkerIsNotRetired(t *testing.T) {
	const sentence = "Water is wet."
	runDir := citedRun(t, "https://water/1")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	gap, _ := markerOfKind(t, runDir, "gap", sentence)
	before := readReport(t, runDir)
	_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "it does not hold", "--anchor", gap)
	if err == nil {
		t.Fatalf("a retire of open gap %s's own marker was recorded", gap)
	}
	for _, w := range []string{"gap " + gap, "open", "cut its sentence"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
		t.Errorf("the refused retire recorded %d retire(s)", n)
	}
	if after := readReport(t, runDir); after != before {
		t.Errorf("the refused retire changed the report:\n%s", after)
	}

	closeQuoteGap(t, runDir, gap)
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "the gap is settled", "--anchor", gap); err != nil {
		t.Fatalf("a retire of closed gap %s's marker was refused: %v", gap, err)
	}
	if rep := readReport(t, runDir); strings.Contains(rep, anchor.Token(gap)) || !strings.Contains(rep, "- "+sentence+"\n") {
		t.Errorf("after the retire the report should hold %q and not the marker:\n%s", sentence, rep)
	}
}
