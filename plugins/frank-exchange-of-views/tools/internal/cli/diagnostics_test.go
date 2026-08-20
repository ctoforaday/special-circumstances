package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func trajectoryFor(t *testing.T, runDir, body string) string {
	t.Helper()
	dir := filepath.Join(filepath.Dir(runDir), ".probe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	ev := map[string]any{"message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "content": body},
	}}}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, filepath.Base(runDir)+".jsonl")
	if err := os.WriteFile(p, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// THE DENOMINATOR IS THE SEAT'S TREE, NOT THE role-PREFIXED JOIN KEYS.
//
// `motion`, `fetch` and `count-claims` are mounted on every seat's root and carry no role in
// their CommandPaths key, so a denominator built from that subset was MISSING them — and because
// a listing containing a name the role does not own is rejected whole, their presence in the
// seat's own help discarded the block. The measure reported a lens that had just been handed its
// entire surface as having seen nothing. Measured, on the first real trajectory.
func TestTheDenominatorIsWhatTheSeatsTreeActuallyOffers(t *testing.T) {
	acts := actsOf("lens")
	for _, want := range []string{"motion", "fetch", "count-claims", "finding"} {
		found := false
		for _, a := range acts {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is on the lens's tree and missing from its act list — a listing carrying it would be rejected whole", want)
		}
	}
}

// A SEAT'S SURFACE COUNTS WHEN IT ARRIVES, however it arrives.
//
// The refusal teaches: a wrong-seat error prints the listing. So exposure is not the same
// question as "did it run --help", and a measure that only counted help reads would score a seat
// that had been shown everything as having seen nothing.
func TestASurfaceThatArrivesThroughARefusalStillCounts(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	trajectoryFor(t, runDir, "Exit code 2\nfeov-record: \"position\" is not on your surface.\n\n"+
		"Available Commands:\n"+
		"  completion  Generate the autocompletion script\n"+
		"  finding     a lens finding\n"+
		"  verify      adjudicate ONE citation\n"+
		"\n")

	out, err := run(t, "show", "diagnostics", "--seat-id", "operator", "--run", runDir, "--sitting", "red-lens-r1-L1")
	if err != nil {
		t.Fatalf("show diagnostics: %v", err)
	}
	var got RunDiagnostic
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("the projection did not emit JSON a reader can join on: %v\n%s", err, out)
	}
	if len(got.Seats) != 1 || got.Seats[0].SeatID != "red-lens-r1-L1" {
		t.Fatalf("want exactly the named sitting, got %+v", got.Seats)
	}
	if got.Seats[0].HelpBlocks != 1 {
		t.Errorf("the listing arrived in a REFUSAL and was not counted: %+v", got.Seats[0])
	}
	if len(got.Seats[0].SeenLeaf) != 2 {
		t.Errorf("SEEN = %v, want the two role verbs", got.Seats[0].SeenLeaf)
	}
}

// AND WITHOUT --sitting IT REFUSES, naming the seats it will not guess between.
//
// A board's BUILD registers every seat so the fixture has parties to hang events on; only one is
// then dispatched. Reporting "the seats that registered" credited three seats with an exposure of
// zero they were never asked for — the plausible zero, produced by the projection built to
// remove it.
func TestWithoutASittingItRefusesRatherThanCreditingSeatsThatNeverSat(t *testing.T) {
	runDir := seatRun(t)
	for _, s := range []string{"red-lens-r1-L1", "red-merge-r1"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", s); err != nil {
			t.Fatal(err)
		}
	}
	trajectoryFor(t, runDir, "nothing here")

	_, err := run(t, "show", "diagnostics", "--seat-id", "operator", "--run", runDir)
	if err == nil {
		t.Fatal("it reported every registered seat; the ones that never sat get an exposure of zero they were never given the chance to earn")
	}
	for _, want := range []string{"red-lens-r1-L1", "red-merge-r1", "--sitting"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name what it will not guess between; missing %q in %v", want, err)
		}
	}
}
