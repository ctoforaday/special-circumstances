package migrate_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// plantInB9 extracts b9 and appends one act to its record in the old schema's own shape — a real
// old-schema record, because a record written by this binary cannot stand in for one written by an
// older one — as the next act in the seat's sequence after the last event b9 recorded.
func plantInB9(t *testing.T, eventType, seat string, stmts ...string) string {
	t.Helper()
	runDir := t.TempDir()
	untar(t, archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"), runDir)
	db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var id int64
	if err := db.QueryRow(`SELECT max(id) + 1 FROM events`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	all := append([]string{`INSERT INTO events (id, seat_id, ts, type, key) VALUES (?, '` + seat + `', '2026-09-11T14:59:00.000000000Z', '` +
		eventType + `', '` + seat + `:` + eventType + `:#2')`}, stmts...)
	for _, stmt := range all {
		if _, err := db.Exec(stmt, id); err != nil {
			t.Fatalf("planting the old act: %v", err)
		}
	}
	return runDir
}

func migratePlanted(t *testing.T, runDir string) (*migration.Manifest, record.Family) {
	t.Helper()
	toDir := recordtest.TmpRun(t)
	m, err := migrate.Migrate(runDir, toDir, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("the planted act did not migrate: %v (manifest: %+v)", err, m)
	}
	fam, err := record.FamilyOf(runtest.Open(t, toDir))
	if err != nil {
		t.Fatal(err)
	}
	return m, fam
}

// AN OLD INQUIRY RULING ARRIVES AS AN INQUIRY RULING. No archived run in this schema's shape carries
// a ruling on what blue proposed in the earliest words for it, so this plants one into b9, which
// wrote them: no motion row, because the proposal is the filing; the subject in those words; the
// ruling in the column named for its arm.
//
// BOTH HALVES MUST MOVE, and they are separate mechanisms. The subject's value is an enum word,
// carried by valueRenames. The ruling itself is a COLUMN named for the oneof arm, which identity
// refuses outright when the current body does not declare it; inquiryArm renames it first. Either
// one missing refuses the event, and this test says which.
func TestAnOldInquiryRulingMigratesUnderTheOneWord(t *testing.T) {
	runDir := plantInB9(t, "motion_rule", "red-chair",
		`INSERT INTO motion_rule (event_id, motion_id, subject, opinion, direction) VALUES (?, 'Q1', 'direction', 'worth the run''s time: trial division to the square root settles primality outright', 'endorsed')`)
	m, fam := migratePlanted(t, runDir)
	found := false
	for _, e := range fam.Events {
		r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || r.GetMotionId() != m.IDs["Q1"] {
			continue
		}
		found = true
		if r.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY {
			t.Errorf("the ruling's subject arrived as %v, want inquiry — the value rename did not run", r.Subject())
		}
		if r.GetInquiry() != recordpb.InquiryRuling_INQUIRY_RULING_ENDORSED {
			t.Errorf("the ruling arrived as %v, want endorsed under the inquiry arm — the column rename did not run", r.GetRuling())
		}
	}
	if !found {
		t.Fatal("the planted ruling on Q1 is not in the migrated record")
	}
}

// AN OLD INQUIRY MOTION ARM ARRIVES AS THE INQUIRY ARM, its id column with it. The old schema kept
// the arm as a table under an earlier name, keyed by an id column under an earlier name; no archive
// holds a row of it, so this plants one into b9. The arm table and its id column are both renamed, or identity refuses
// the act.
func TestAnOldInquiryMotionArmMigratesUnderTheOneWord(t *testing.T) {
	runDir := plantInB9(t, "motion", "red-chair",
		`INSERT INTO motion (event_id, motion_id, subject, basis, filing_case) VALUES (?, 'M9', 'direction', 'the proposal is in scope and its hypothesis carries its budget', 'direction')`,
		`INSERT INTO motion_direction (event_id, avenue_id) VALUES (?, 'Q1')`)
	m, fam := migratePlanted(t, runDir)
	found := false
	for _, e := range fam.Events {
		mo, ok := recordpb.BodyAs[*recordpb.Motion](e)
		if !ok || mo.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY {
			continue
		}
		found = true
		if got := mo.GetInquiry().GetInquiryId(); got == "" || got != m.IDs["Q1"] {
			t.Errorf("the motion's inquiry arm arrived holding %q, want the migrated id of Q1 (%q)", got, m.IDs["Q1"])
		}
	}
	if !found {
		t.Fatal("the planted motion under subject inquiry is not in the migrated record")
	}
}
