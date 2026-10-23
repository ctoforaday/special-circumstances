package migrate_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// statedEpoch is the event-schema epoch an archived record states, asked of the file with plain SQL
// so the answer does not pass through the open under test. ok is false for a run that holds no
// database, or one whose database states no epoch.
func statedEpoch(t *testing.T, runDir string) (epoch int, ok bool) {
	t.Helper()
	path := filepath.Join(runDir, "records", "record.db")
	if _, err := os.Stat(path); err != nil {
		return 0, false
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.QueryRow(`SELECT "epoch" FROM "schema_epoch" WHERE "id" = 1`).Scan(&epoch); err != nil {
		return 0, false
	}
	return epoch, true
}

// EVERY ARCHIVED RUN BEHIND THIS BINARY IS REFUSED AT THE READ, NAMING MIGRATE. An archive is the
// one database an older binary really wrote, and the open is the only place that says so: no reader
// past it asks again whether a table or a column is there. So a record the open lets through is read
// as this binary's shape, and this is the census that none of an older shape gets through — beside
// TestEveryArchivedRunMigrates, which is the other half of the same promise.
func TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	refused := 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			runDir := recordtest.RequireArchive(t, name)
			epoch, stated := statedEpoch(t, runDir)
			run, err := record.NewRun(runDir)
			if err != nil {
				t.Fatal(err)
			}
			_, err = record.FamilyOf(run)
			if err != nil {
				refused++
				if !strings.Contains(err.Error(), "`feov-migrate --from <runDir> --to <freshDir>`") {
					t.Errorf("%s is refused without naming the way out: %v", name, err)
				}
				return
			}
			if !stated || epoch != record.EventSchema {
				t.Errorf("%s states epoch %d (stated: %v), this binary writes %d, and the record was read without a refusal",
					name, epoch, stated, record.EventSchema)
			}
		})
	}
	// Every archive reading clean would make the loop above a test of nothing.
	if refused == 0 {
		t.Errorf("none of %d archived runs was refused — no record behind this binary was offered to the open", len(tarballs))
	}
}
