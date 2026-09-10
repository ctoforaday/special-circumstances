package catalogue

import (
	"database/sql"
	"io"
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
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
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

// A REBUILD IS PENDING UNTIL A BACKFILL FOLLOWS IT, and the comparison is strict: a backfill that
// completes in the rebuild's own second did read the corpus after it. An unreadable marker is an
// ERROR — never "not pending", which would fold "cannot tell" into a clean board.
func TestRebuildPending(t *testing.T) {
	rebuilt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name     string
		set      func(t *testing.T, db *sql.DB)
		pending  bool
		from     int
		wantsErr bool
	}{
		{name: "no markers", set: func(*testing.T, *sql.DB) {}},
		{name: "rebuilt only", set: func(t *testing.T, db *sql.DB) {
			if err := MarkRebuilt(db, 2, rebuilt); err != nil {
				t.Fatal(err)
			}
		}, pending: true, from: 2},
		{name: "backfilled after", set: func(t *testing.T, db *sql.DB) {
			MarkRebuilt(db, 2, rebuilt)
			MarkBackfilled(db, rebuilt.Add(time.Second))
		}, from: 2},
		{name: "backfilled before", set: func(t *testing.T, db *sql.DB) {
			MarkRebuilt(db, 1, rebuilt)
			MarkBackfilled(db, rebuilt.Add(-time.Second))
		}, pending: true, from: 1},
		{name: "backfilled in the same second", set: func(t *testing.T, db *sql.DB) {
			MarkRebuilt(db, 2, rebuilt)
			MarkBackfilled(db, rebuilt.Add(900*time.Millisecond))
		}, from: 2},
		{name: "corrupt rebuilt_at", set: func(t *testing.T, db *sql.DB) {
			MarkRebuilt(db, 2, rebuilt)
			mustExec(t, db, `UPDATE meta SET value='yesterday' WHERE key='rebuilt_at'`)
		}, wantsErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			tc.set(t, db)
			at, from, pending, err := RebuildPending(db)
			if tc.wantsErr {
				if err == nil {
					t.Fatalf("a corrupt marker read as pending=%v with no error", pending)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if pending != tc.pending {
				t.Errorf("pending = %v, want %v", pending, tc.pending)
			}
			if pending && (from != tc.from || !at.Equal(rebuilt)) {
				t.Errorf("rebuilt at %v from %d, want %v from %d", at, from, rebuilt, tc.from)
			}
		})
	}
}
