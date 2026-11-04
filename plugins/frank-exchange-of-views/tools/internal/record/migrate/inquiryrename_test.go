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

// plantedDropped is the planted act's own accounting: the manifest counts it in under its word and
// writes nothing for it, so the migrated record holds exactly the events b9 alone migrates to.
func plantedDropped(t *testing.T, m *migration.Manifest, word string) {
	t.Helper()
	base, _ := migrate.MigratedArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	if m.In[word] != base.In[word]+1 {
		t.Fatalf("the manifest counts %d %s event(s) in, b9 alone %d: the planted act was not read", m.In[word], word, base.In[word])
	}
	if m.Out[word] != base.Out[word] {
		t.Errorf("the planted act arrived as an event: %d %s event(s) out, b9 alone %d", m.Out[word], word, base.Out[word])
	}
}

// AN OLD RULING ON AN INQUIRY TRANSLATES TO NO EVENT. No archived run in this schema's shape carries
// a ruling on what blue proposed in the earliest words for it, so this plants one into b9, which
// wrote them: the subject in those words, the ruling in the column named for its arm. inquiryArm
// renames the column to `inquiry`, and a ruling holding that column has no equivalent; without the
// translation the column is one the current ruling does not declare, and the act refuses.
func TestAnOldInquiryRulingTranslatesToNoEvent(t *testing.T) {
	runDir := plantInB9(t, "motion_rule", "red-chair",
		`INSERT INTO motion_rule (event_id, motion_id, subject, opinion, direction) VALUES (?, 'Q1', 'direction', 'worth the run''s time: trial division to the square root settles primality outright', 'endorsed')`)
	m, fam := migratePlanted(t, runDir)
	plantedDropped(t, m, "motion_rule")
	for _, e := range fam.Events {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetMotionId() == m.IDs["Q1"] {
			t.Errorf("the planted ruling on Q1 arrived as an event: %v", r)
		}
	}
}

// AN OLD MOTION ON AN INQUIRY TRANSLATES TO NO EVENT. The old schema kept the arm as a table under an
// earlier name, keyed by an id column under an earlier name; no archive holds a row of it, so this
// plants one into b9, filed by blue. inquiryArm renames the arm to `inquiry`, and a motion filed
// under it has no equivalent; without the translation the arm is one the current motion does not
// declare, and the act refuses.
func TestAnOldInquiryMotionTranslatesToNoEvent(t *testing.T) {
	runDir := plantInB9(t, "motion", "blue-respond",
		`INSERT INTO motion (event_id, motion_id, subject, basis, filing_case) VALUES (?, 'M9', 'direction', 'the proposal is in scope and its hypothesis carries its budget', 'direction')`,
		`INSERT INTO motion_direction (event_id, avenue_id) VALUES (?, 'Q1')`)
	m, fam := migratePlanted(t, runDir)
	plantedDropped(t, m, "motion")
	for _, e := range fam.Events {
		if mo, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && mo.GetMotionId() == m.IDs["M9"] && m.IDs["M9"] != "" {
			t.Errorf("the planted motion on Q1 arrived as an event: %v", mo)
		}
	}
}
