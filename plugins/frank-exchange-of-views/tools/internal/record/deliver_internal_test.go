package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// deliver is deliverytest.Deliver for this package's own tests, which cannot import it: seatID's
// delivery of report text in its open sitting, cut and recorded as the delivery writer does.
func deliver(t testing.TB, run Run, seatID string) int64 {
	t.Helper()
	w, err := WorkOfSeat(run, seatclass.RoleOfSeat(seatID), seatID)
	if err != nil {
		t.Fatalf("reading %s's work list to cut its delivery: %v", seatID, err)
	}
	cut, err := CutDelivery(run, seatID, ReportTextAreas(w), func(string, string) bool { return true })
	if err != nil {
		t.Fatalf("cutting %s's delivery: %v", seatID, err)
	}
	if len(cut.Pieces) == 0 {
		t.Fatalf("%s is owed no report text: the cut holds none (no report yet, or nothing its role is handed)", seatID)
	}
	if err := RecordDelivery(run, seatID, cut.Head); err != nil {
		t.Fatalf("recording %s's delivery at head %d: %v", seatID, cut.Head, err)
	}
	return cut.Head
}
