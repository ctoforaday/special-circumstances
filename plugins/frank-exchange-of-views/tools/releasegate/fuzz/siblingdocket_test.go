package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE BENCH RULES EVERY DOCKET MOTION STANDING ON A GAP IT IS ENGAGED ON — pinned deterministically,
// because the sweep reaches it on about one forced-VERIFIED run in twenty-five.
//
// A party may file a docket motion on a gap whose motion already stands unruled, and the record
// admits it. Each motion is a blocker on the bench's work list, and the plan keys the bench's sitting
// on the newest filing, so a bench that rules one of the two has sat for the other and ruled nothing:
// it is never readied again, and the run ends UNVERIFIED with nobody ready. The fuzz bench ruled
// the first motion on each engaged gap and left the sibling standing.
func TestTheBenchRulesEverySiblingDocketMotion(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"),
		[]byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(7))
	r.register("blue", "blue-synthesize")
	if _, err := r.do("ingest", "blue-synthesize").result(); err != nil {
		t.Fatalf("ingest the round-0 report: %v", err)
	}
	r.register("chair", "red-chair")
	lens := fuzzLensSeats[0]
	r.register("lens", lens)
	gapID := r.mint(lens)
	if gapID == "" {
		t.Fatalf("the mint was refused: %s", r.actErr)
	}
	r.register("blue", "blue-respond")
	for i := 0; i < 2; i++ {
		if _, ok := r.admit("motion", "docket", "file", "--seat-id", "blue-respond", "--id", gapID,
			"--reason", "fuzz: "+gapID+" is the bench's to settle"); !ok {
			t.Fatalf("docket motion %d on %s was refused: %s", i+1, gapID, r.actErr)
		}
	}
	if got := r.unruledDocketMotions(gapID); len(got) != 2 {
		t.Fatalf("the setup stands %d unruled docket motion(s) on %s, want the two filed: %v", len(got), gapID, got)
	}

	r.sit("bench", "judge", "docket")
	r.benchDisposes("judge", gapID, "remanded", "--direction", "blue pursuing the stated direction",
		"--reason", "fuzz: carried with a stated direction for "+gapID)
	if r.actErr != "" {
		t.Fatalf("the bench's rulings: %s", r.actErr)
	}
	if left := r.unruledDocketMotions(gapID); len(left) > 0 {
		t.Errorf("the bench sat for %s and left docket motion(s) %v standing — the plan never readies it for them again", gapID, left)
	}
}
