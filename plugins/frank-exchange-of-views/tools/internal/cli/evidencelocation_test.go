package cli

import (
	"encoding/json"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A SOURCE'S ROW IS LOCATED AS THE BOARD LOCATES A QUOTE GAP (D13, #1727; T29). `show evidence`
// prints the sentence around the citation's token in the report AS IT STANDS under `marked`, so a
// reworded sentence reads reworded and never as the write-time text; the sentence as cited under
// `gone` once the citation is retired; and a quote gap on the same sentence locates to the same
// words, since both read one function (record.locateAnchor).
func TestShowEvidenceLocatesACitationAsTheBoardLocatesAGap(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, id := range []string{"red-chair", lensSeat, "blue-respond"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", id); err != nil {
			t.Fatal(err)
		}
	}
	seedBlueReport(t, runDir)
	const claim = "§2 the finding prose lands in a quoted sentence."
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://example.org/red": []byte("<html>the source</html>")}})
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", "blue-respond",
		"--url", "https://example.org/red", "--title", "T", "--quote", claim); err != nil {
		t.Fatal(err)
	}
	source := func() record.EvidenceSourceJSON {
		t.Helper()
		out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--json")
		if err != nil {
			t.Fatalf("show evidence --json: %v", err)
		}
		var ev record.EvidenceJSON
		if err := json.Unmarshal([]byte(out), &ev); err != nil {
			t.Fatalf("show evidence --json: %v\n%s", err, out)
		}
		if len(ev.Sources) != 1 {
			t.Fatalf("sources = %+v, want one row", ev.Sources)
		}
		return ev.Sources[0]
	}
	if s := source(); s.LocationState != record.LocationMarked || s.Location != claim {
		t.Fatalf("after the cite: location = %q %q, want `marked` with the cited sentence", s.LocationState, s.Location)
	}

	// BLUE REWORDS THE SENTENCE: the row reads the sentence as it now stands, not the write-time text.
	printed, tok := anchoredLine(t, runDir)
	const reworded = "§2 an entirely different assertion now"
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", printed, "--new", reworded+tok+".", "--reason", "rewrote the cited sentence"); err != nil {
		t.Fatal(err)
	}
	s := source()
	if s.LocationState != record.LocationMarked || s.Location != reworded+"." {
		t.Fatalf("after the edit: location = %q %q, want `marked` with the reworded sentence", s.LocationState, s.Location)
	}

	// A QUOTE GAP ON THE SAME SENTENCE LOCATES TO THE SAME WORDS: one function locates both.
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--quote", reworded+".", "--class", "spec_underspecification",
		"--problem", "the defect", "--fix", "the fix", "--check-kind", "document", "--check", "the check",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low"); err != nil {
		t.Fatalf("mint on the sentence: %v", err)
	}
	gap := lastBody(t, runDir, &recordpb.Mint{}).GetGapId()
	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatal(err)
	}
	var board record.BoardJSON
	if err := json.Unmarshal([]byte(out), &board); err != nil {
		t.Fatalf("show board --json: %v\n%s", err, out)
	}
	var located bool
	for _, g := range board.Open {
		if g.ID == gap {
			located = true
			if g.LocationState != s.LocationState || g.Location != s.Location {
				t.Errorf("the board locates the gap at %q %q and the evidence view its citation at %q %q — one sentence, two locations", g.LocationState, g.Location, s.LocationState, s.Location)
			}
		}
	}
	if !located {
		t.Fatalf("the board does not hold gap %s:\n%s", gap, out)
	}

	// BLUE RETIRES THE CITATION: the row is `gone`, with the sentence as cited.
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", "blue-respond", "--anchor", s.Anchor, "--reason", "it does not hold"); err != nil {
		t.Fatalf("retire: %v", err)
	}
	out, err = run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--anchor", s.Anchor, "--json")
	if err != nil {
		t.Fatalf("show evidence --anchor on the retired citation: %v", err)
	}
	var at record.EvidenceAtJSON
	if err := json.Unmarshal([]byte(out), &at); err != nil {
		t.Fatalf("show evidence --anchor --json: %v\n%s", err, out)
	}
	if at.Source == nil || at.Source.LocationState != record.LocationGone || at.Source.Location != claim {
		t.Errorf("after the retire: %+v, want `gone` with the sentence as cited", at.Source)
	}
}
