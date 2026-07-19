package record

import (
	"testing"
	"time"
)

// THE ORDERING KEY MUST BE FINE ENOUGH TO ORDER.
//
// `ts` is not a field a human reads; it is the key replay sorts by, and when two events
// share it the sort falls through to (SeatID, Seq) — ordering by SEAT NAME, which is the
// exact defect that silently dropped the bench's closures ("judge-r2" sorting before
// "red-merge-r1", so every ruling replayed before the mint it referenced).
//
// The stamp was millisecond-precision, so any two events inside one tick tied and fell
// back to that broken order.
//
// THE CLOCK IS INJECTED, deliberately. The first version of this test appended twenty
// events back-to-back and asserted the stamps were distinct — and it PASSED on the
// millisecond clock it was written to condemn, because file I/O makes each append take
// longer than a tick on this machine. It proved the machine was slow, not that the format
// was adequate, and it would have gone green on the very build CI rejected. Driving Now
// with a fake clock that advances by less than a millisecond asks the real question:
// can this format represent two events that happen close together?
func TestEventStampsResolveSubMillisecondEvents(t *testing.T) {
	orig := Now
	defer func() { Now = orig }()

	base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	var tick int
	// 100 microseconds apart: far coarser than the clock actually delivers, and still
	// invisible to a millisecond format.
	Now = func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * 100 * time.Microsecond)
	}

	const n = 20
	seen := map[string]bool{}
	var stamps []string
	for i := 0; i < n; i++ {
		s := stamp()
		seen[s] = true
		stamps = append(stamps, s)
	}

	if len(seen) != n {
		t.Errorf("%d events 100µs apart produced only %d distinct stamps — the %d collisions fall back to ordering by SEAT NAME, which is how the bench's closures were dropped",
			n, len(seen), n-len(seen))
	}
	for i := 1; i < len(stamps); i++ {
		if stamps[i] <= stamps[i-1] {
			t.Errorf("stamps are not strictly increasing: %s then %s. Lexicographic order must BE time order, or the fixed-width format is lying about what it encodes", stamps[i-1], stamps[i])
		}
	}
}

// And the whole path, end to end: an appended event carries a stamp at all.
func TestAppendedEventCarriesAStamp(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	ev, err := Append(runDir, "red-lens-r1-L1", "observe", NewPayload().Set("label", "L1-O1"))
	if err != nil {
		t.Fatal(err)
	}
	if ev.TS == "" {
		t.Fatal("an event carries no timestamp; replay would order it by seat name")
	}
}
