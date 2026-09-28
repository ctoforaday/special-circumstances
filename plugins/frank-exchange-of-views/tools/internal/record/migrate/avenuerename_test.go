package migrate_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN AVENUE RULING IN AN OLD RECORD ARRIVES AS AN AVENUE RULING, and this is the one rename of the
// collapse that no archive exercises on its own.
//
// The concept has one word now. The record called what blue proposes and the chair rules on a
// `direction` in two places — the motion subject's value, and the ruling's own column — and the
// surface called it an `inquiry`. Every archive migrates the review of the set (`inquiry_review`,
// six rows in b9) and the anchor kind (`about_kind = inquiry`, in b, b6 and b7), so those renames are
// exercised by the archive tests already. But not one archived run carries an avenue RULING: the
// chair never ruled on one in any captured run. So this plants one — into a real old-schema record,
// because a record written by this binary cannot stand in for one written by an older one — and
// migrates it.
//
// BOTH HALVES MUST MOVE, and they are separate mechanisms. The subject's value `direction` is an
// enum word, carried by valueRenames. The ruling itself is a COLUMN named for the oneof arm, which
// identity refuses outright when the current body does not declare it; avenueArm renames it first.
// Either one missing refuses the event, and this test says which.
func TestAnOldAvenueRulingMigratesUnderTheOneWord(t *testing.T) {
	runDir := t.TempDir()
	untar(t, archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"), runDir)

	db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	// The next act in the chair's own sequence, after the last event b9 recorded, ruling on the
	// avenue b9 proposed first — exactly the shape the old schema wrote: no motion row, because the
	// proposal is the filing; the ruling in the column named for its arm.
	var id int64
	if err := db.QueryRow(`SELECT max(id) + 1 FROM events`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`INSERT INTO events (id, seat_id, ts, type, key) VALUES (?, 'red-chair', '2026-09-11T14:59:00.000000000Z', 'motion_rule', 'red-chair:motion_rule:#2')`,
		`INSERT INTO motion_rule (event_id, motion_id, subject, opinion, direction) VALUES (?, 'Q1', 'direction', 'worth the run''s time: trial division to the square root settles primality outright', 'endorsed')`,
	} {
		if _, err := db.Exec(stmt, id); err != nil {
			t.Fatalf("planting the old ruling: %v", err)
		}
	}
	db.Close()

	toDir := recordtest.TmpRun(t)
	m, err := migrate.Migrate(runDir, toDir, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("an old avenue ruling did not migrate: %v (refusals: %+v)", err, m)
	}

	fam, err := record.FamilyOf(runtest.Open(t, toDir))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range fam.Events {
		r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || r.GetMotionId() != "Q1" {
			continue
		}
		found = true
		if r.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_AVENUE {
			t.Errorf("the ruling's subject arrived as %v, want avenue — the value rename did not run", r.GetSubject())
		}
		if r.GetAvenue() != recordpb.AvenueRuling_AVENUE_RULING_ENDORSED {
			t.Errorf("the ruling arrived as %v, want endorsed under the avenue arm — the column rename did not run", r.GetRuling())
		}
	}
	if !found {
		t.Fatal("the planted ruling on Q1 is not in the migrated record")
	}
}
