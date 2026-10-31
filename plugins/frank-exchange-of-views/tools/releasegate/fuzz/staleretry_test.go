package fuzz

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// staleMinter mints the gap staleRun's lenses file on; staleFiler is the lens whose finding about it
// meets the stale refusals; staleSiblings sit beside it and each lands one finding on the gap.
const (
	staleMinter = "red-lens-logic"
	staleFiler  = "red-lens-dark-side"
)

var staleSiblings = []string{"red-lens-adversary", "red-lens-architecture", "red-lens-computation", "red-lens-evidence"}

// staleRun is a record holding one gap staleMinter minted, with every lens of the cast seated.
func staleRun(t *testing.T) (*runner, string) {
	t.Helper()
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	const sentence = "A § fuzz sentence to anchor findings."
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte("# § fuzz\n\n"+sentence+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
		[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":1,"k":1,"kMax":2,"mintBudget":1,"convergenceFraction":0.25,"maxEpochs":10}`)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	cast := append([]string{staleMinter, staleFiler}, staleSiblings...)
	cast = append(cast, "red-chair", "blue-synthesize", "blue-respond", "judge")
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, &lockedRand{r: rand.New(steadySource(99))})
	r.register("blue", "blue-synthesize")
	if _, err := r.do("ingest", "blue-synthesize").result(); err != nil {
		t.Fatalf("ingest the report: %v", err)
	}
	r.chairPlans(t)
	r.dispatchSits("lens", staleMinter)
	r.mustExec(t, "--json", "--seat-id", staleMinter, "show", "report")
	r.mustExec(t, "--json", "--seat-id", staleMinter, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	r.dispatchReturns(staleMinter, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	return r, open[0]
}

// siblingFinds is a lens sitting beside the filer and landing one finding about gapID.
func (r *runner) siblingFinds(t *testing.T, seatID, gapID string) {
	t.Helper()
	r.dispatchSits("lens", seatID)
	r.mustExec(t, "--json", "finding", "--seat-id", seatID, "--about-kind", "gap", "--about", gapID,
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--reason", "fuzz: "+seatID+" on the same gap")
	r.dispatchReturns(seatID, false)
}

// A LENS'S ACT ON A GAP ITS SIBLINGS KEEP CHANGING LANDS ONCE THEY STOP (seed 9 of #1660's sweep):
// the filer sits, a sibling files about the gap, and the filer's finding is refused as a stale read;
// between each of its attempts another sibling lands a finding on the gap, so each attempt is refused
// for a change the previous refusal could not show. The record is right each time, and the act lands
// at the first attempt no sibling's act preceded. Four siblings is more refusals than the three
// attempts the drive once allowed.
func TestAnActRefusedForEachSiblingsChangeLandsOnceTheyStop(t *testing.T) {
	r, gap := staleRun(t)
	r.dispatchSits("lens", staleFiler)
	r.siblingFinds(t, staleSiblings[0], gap)
	next := 1
	r.betweenStaleRetries = func(seatID, gapID string) {
		if seatID != staleFiler || gapID != gap {
			t.Errorf("a retry of %s on %s, want %s's on %s", seatID, gapID, staleFiler, gap)
		}
		if next < len(staleSiblings) {
			r.siblingFinds(t, staleSiblings[next], gap)
			next++
		}
	}
	if _, ok := r.admitFresh(gap, "finding", "--seat-id", staleFiler, "--about-kind", "gap", "--about", gap,
		"--severity", "medium", "--likelihood", "low", "--impact", "medium", "--reason", "fuzz finding"); !ok {
		t.Fatalf("the filer's finding was not admitted: %s", r.actErr)
	}
	r.dispatchReturns(staleFiler, false)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if got := r.staleRefusals(staleFiler, gap); got != len(staleSiblings) {
		t.Errorf("the filer met %d stale refusal(s) on %s, want one per sibling's finding, %d", got, gap, len(staleSiblings))
	}
}

// A STALE REFUSAL THAT MOVES THE READ POINT TO NO OTHER SEAT'S ACT ENDS THE RETRIES: retried on it,
// the act would be bounded by nothing the other seats do. A Stale event through such an event stands
// in for the refusal, which the record never writes: through the filer's own act, and through a
// sibling's Stale event, which is a refusal and lands nothing.
func TestAReadPointMovedToNoOtherSeatsChangeIsNoRetry(t *testing.T) {
	for _, c := range []struct {
		name  string
		lands func(t *testing.T, r *runner, gap string)
	}{
		{"the filer's own act", func(t *testing.T, r *runner, _ string) {
			r.mustExec(t, "--json", "--seat-id", staleFiler, "show", "report")
		}},
		{"a sibling's refusal", func(t *testing.T, r *runner, gap string) {
			b, ok := r.board()
			if !ok {
				t.Fatal("the board could not be read")
			}
			at := b.At.Of(b.Events[len(b.Events)-1]).ID
			if _, err := record.Append(record.Identity{Run: r.run(), SeatID: staleSiblings[0]}, &recordpb.Stale{GapId: &gap, Through: &at}); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, gap := staleRun(t)
			r.dispatchSits("lens", staleFiler)
			before := r.readPointOf(staleFiler, gap)
			c.lands(t, r, gap)
			b, ok := r.board()
			if !ok {
				t.Fatal("the board could not be read")
			}
			at := b.At.Of(b.Events[len(b.Events)-1]).ID
			if _, err := record.Append(record.Identity{Run: r.run(), SeatID: staleFiler}, &recordpb.Stale{GapId: &gap, Through: &at}); err != nil {
				t.Fatal(err)
			}
			if r.staleMoved(staleFiler, gap, before) {
				t.Errorf("a read point moved to event %d, %s, was taken for another seat's change", at, c.name)
			}
			if r.actErr == "" {
				t.Errorf("the drive noted nothing for a read point moved to %s", c.name)
			}
			r.dispatchReturns(staleFiler, false)
		})
	}
}

// staleRefusals counts the Stale events seatID has on gapID.
func (r *runner) staleRefusals(seatID, gapID string) int {
	b, ok := r.board()
	if !ok {
		return 0
	}
	n := 0
	for _, e := range b.Events {
		if st, ok := recordpb.BodyAs[*recordpb.Stale](e); ok && e.GetSeatId() == seatID && st.GetGapId() == gapID {
			n++
		}
	}
	return n
}
