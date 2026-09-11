package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CORRECTED CITE IS THE CITE, ONCE, WITH ITS NEW TITLE. The replacement carries the original's
// label and quote, so the report places the marker once where it always stood, and the
// Bibliography reads the title that stands — not the struck one first-seen by label. B8's
// synthesizer left a flagged proof note in the report because re-proving would duplicate the anchor.
func TestACorrectedCiteRendersOnceWithItsNewTitle(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/c": []byte("<html>a source</html>")}})
	quote := "§2 the finding prose lands in a quoted sentence."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", quote, "--url", "https://src/c", "--title", "Source, as read in this run"})
	must(t, runDir, "cite", "--seat-id", "blue-respond", "--quote", quote, "--url", "https://src/c",
		"--title", "Source, section 2", "--corrects", k, "--correction-why", "the title narrated the run")

	if n := len(citeAnchorRe.FindAllString(readReport(t, runDir), -1)); n != 1 {
		t.Fatalf("a corrected cite placed %d markers, want 1:\n%s", n, readReport(t, runDir))
	}
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Title != "Source, section 2" {
		t.Fatalf("the sources = %+v, want one carrying the corrected title", srcs)
	}

	// What identifies the source is not the seat's wording: a correction that moves it is refused.
	k2 := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", "§1 first — a finding sits in sec 1 here.", "--url", "https://src/c", "--title", "Another"})
	_, err = run(t, "cite", "--run", runDir, "--seat-id", "blue-respond", "--quote", "§1 first — a finding sits in sec 1 here.",
		"--url", "https://src/elsewhere", "--title", "Another", "--corrects", k2, "--correction-why", "wrong source")
	if err == nil || !strings.Contains(err.Error(), "may change only the seat's own wording") {
		t.Fatalf("a correction moving the url was not refused as a frozen field: %v", err)
	}
}

// THE REPLACEMENT REPLAYS WHERE ITS ORIGINAL STOOD. The seat edits the sentence its cite anchors
// — carrying the marker, as an edit must — and then corrects the cite. The correction carries the
// original quote, which the report no longer reads; replayed at its own later place it would look
// for that quote after the edit and find nothing. Replayed at its original's place, it lands before
// the edit, and the edit carries it as it carried the original.
func TestACiteCorrectedAfterAnEditToItsSentenceRendersOnce(t *testing.T) {
	runDir := corrFixture(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://src/e": []byte("<html>a source</html>")}})
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"cite", "--quote", "§2 the finding prose lands in a quoted sentence.", "--url", "https://src/e", "--title", "Source, as read in this run"})
	marker := citeAnchorRe.FindString(readReport(t, runDir))
	if marker == "" {
		t.Fatalf("the cite placed no marker:\n%s", readReport(t, runDir))
	}
	must(t, runDir, "edit", "--seat-id", "blue-respond", "--quote", "lands in a quoted sentence"+marker,
		"--new", "rests in a quoted sentence"+marker, "--reason", "a plainer verb")
	must(t, runDir, "cite", "--seat-id", "blue-respond", "--title", "Source, section 2",
		"--corrects", k, "--correction-why", "the title narrated the run")

	report := readReport(t, runDir)
	if n := len(citeAnchorRe.FindAllString(report, -1)); n != 1 || !strings.Contains(report, "rests in a quoted sentence"+marker) {
		t.Fatalf("after an edit and a correction the report carries %d markers, want one on the edited sentence:\n%s", n, report)
	}
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Title != "Source, section 2" {
		t.Fatalf("the sources = %+v, want one carrying the corrected title", srcs)
	}
}

// A CORRECTED PROOF IS THE PROOF, ONCE, WITH ITS NEW NOTE — and it does not run again.
func TestACorrectedProofRendersOnceWithItsNewNote(t *testing.T) {
	runDir := corrFixture(t)
	s := script(t, runDir, "p.js", "console.log(91 % 7)")
	quote := "the parser accepts an empty body in this line."
	k := correctionKeyOf(t, runDir, "blue-respond", []string{"prove", "--quote", quote, "--script", s, "--reason", "this run checked it"})
	must(t, runDir, "prove", "--seat-id", "blue-respond", "--quote", quote, "--script", s,
		"--reason", "91 leaves remainder 0 on division by 7", "--corrects", k, "--correction-why", "the note narrated the run")

	if n := strings.Count(readReport(t, runDir), "<!--proof:"); n != 1 {
		t.Fatalf("a corrected proof placed %d markers, want 1:\n%s", n, readReport(t, runDir))
	}
	proofs, err := record.RecordedProofs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(proofs) != 1 || proofs[0].Reason != "91 leaves remainder 0 on division by 7" {
		t.Fatalf("the proofs = %+v, want one carrying the corrected note", proofs)
	}
}
