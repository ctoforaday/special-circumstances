package catalogue

import (
	"path/filepath"
	"testing"
	"time"
)

// RETENTION RUNS ONCE A DAY, AND THE FIRST RUN COUNTS.
//
// The interesting case is the first: an empty store has no marker, and "no marker" must read as
// DUE, not as "already done today". Getting that backwards gives a brand-new store that never
// prunes until its second calendar day — a bug whose only symptom is a store slightly too big,
// which is exactly the kind nobody reports.
func TestRetentionIsDueOncePerUTCDay(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	day1 := time.Date(2026, 9, 9, 23, 30, 0, 0, time.UTC)
	if !DueForRetention(db, day1) {
		t.Fatal("a store that has never retained must be due")
	}
	if err := MarkRetained(db, day1); err != nil {
		t.Fatal(err)
	}
	if DueForRetention(db, day1.Add(20*time.Minute)) {
		t.Error("retention ran twice in one UTC day")
	}
	// 00:10 the next day: a new UTC date, and only forty minutes later. The boundary is the
	// DATE, not a rolling 24 hours, and this is the case that tells the two apart.
	if !DueForRetention(db, day1.Add(40*time.Minute)) {
		t.Error("a new UTC date did not re-arm retention")
	}
	// A local clock east of UTC would already be on the next date at 23:30; the marker must not
	// follow it, or the same store retains twice on a machine that moves.
	east := time.FixedZone("east", 10*3600)
	if DueForRetention(db, day1.In(east)) {
		t.Error("the marker followed the local date instead of UTC")
	}
}
