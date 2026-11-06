package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// EVERY ITEM ON EVERY PROBE BOARD SAYS WHAT IT IS, EVERY OWED ITEM IS QUEUED, AND EVERY LENS DELIVERY
// FITS. Over each board's record, for every fixed seat: each item carries a kind (record.ItemKinds), an
// item the seat owes (SittingJSON.Owed) is on its queue (record.QueueOf) — an owed item off the queue
// is never handed to a lens, and its envelope is refused over it — and a lens's delivery at dispatch
// is within sittingwrite.DeliveryBudget, below the harness's cap past which it reaches the seat as a
// file's preview.
func TestEveryProbeBoardsListIsKeyedQueuedAndFits(t *testing.T) {
	withCLIEngine(t, func(int) (string, error) { return seatprobe.ScannedFixtureSentence, nil })
	kinds := map[record.ItemKind]bool{}
	for _, k := range record.ItemKinds() {
		kinds[k] = true
	}
	max := 0
	for name, b := range seatprobe.Boards() {
		runDir := recordtest.TmpRun(t)
		t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
		exec := func(args ...string) (string, error) { return run(t, args...) }
		r := runtest.Open(t, runDir)
		if err := seatprobe.Build(r, b, exec); err != nil {
			t.Fatalf("board %q does not build: %v", name, err)
		}
		for seat := range seatclass.Seats {
			role := seatclass.RoleOfSeat(seat)
			if role == seatclass.OperatorRole {
				continue
			}
			w, err := record.WorkOfSeat(r, role, seat)
			if err != nil {
				t.Fatalf("board %q, %s: %v", name, seat, err)
			}
			for _, it := range w.Sitting.Open {
				if !kinds[it.Kind] {
					t.Errorf("board %q, %s: item %q carries kind %q, which is no item kind", name, seat, it.What, it.Kind)
				}
			}
			owed, _ := w.Sitting.Owed()
			q := record.QueueOf(w)
		owed:
			for _, it := range owed {
				for _, e := range q {
					for _, qi := range e.Items {
						if qi.What == it.What {
							continue owed
						}
					}
				}
				t.Errorf("board %q, %s: owed item %q is not queued", name, seat, it.What)
			}
			if role != "lens" {
				continue
			}
			n := len(sittingwrite.RenderLensDelivery(w))
			if n > max {
				max = n
			}
			if n > sittingwrite.DeliveryBudget {
				t.Errorf("board %q, %s: the delivery is %d bytes, over the budget of %d", name, seat, n, sittingwrite.DeliveryBudget)
			}
		}
	}
	t.Logf("largest lens delivery over the probe boards: %d bytes (budget %d)", max, sittingwrite.DeliveryBudget)
}
