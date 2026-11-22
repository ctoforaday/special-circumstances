package fuzz

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// staleMinter mints the gap staleRun's lenses act on; staleFiler is the lens whose act on it meets
// the stale refusals; staleSiblings sit beside it.
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
	const sentence = "A § fuzz sentence to anchor findings."
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
		fuzzRunConfig(1, 1, 2, 10)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	cast := append([]string{staleMinter, staleFiler}, staleSiblings...)
	cast = append(cast, "red-chair", "blue-lane-1", "blue-synthesize", "blue-respond", "judge")
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: []string{"blue-lane-1"}}); err != nil {
		t.Fatal(err)
	}
	recordtest.SettleQuestion(t, stageRun.Dir()) // the question is not this fixture's subject
	r := newRunner(bin, runDir, &lockedRand{r: rand.New(steadySource(99))})
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
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

// steadySource draws n on every Intn(100): its 31-bit draw is n, below every rejection bound.
type steadySource int64

func (s steadySource) Int63() int64 { return int64(s) << 32 }
func (steadySource) Seed(int64)     {}
