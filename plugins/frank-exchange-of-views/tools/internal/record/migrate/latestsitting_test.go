package migrate_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE LATEST SITTING IS THE VIEW'S, ON EVERY ARCHIVED RUN. recordsql.LatestSitting finds a seat's
// latest opening through two indexes — its opening registers, and the brackets naming it — where
// the `sittings` view says whose sitting each opening is by joining every event to its bracket. The
// view is the oracle: for every seat on every migrated run, and for a seat that never sat, the
// lookup must name the view's latest sitting and the rank stored on it. Some run must have a seat
// whose latest sitting a bracket opened, or the bracket half was never compared.
func TestTheLatestSittingIsTheViewsOnEveryArchivedRun(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	byBracket, seats := 0, 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			_, runDir := migrate.MigratedArchive(t, name)
			db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rows, err := db.Query(`SELECT "seat_id" FROM "events" UNION SELECT "seat_id" FROM "sittings" UNION SELECT 'a-seat-that-never-sat'`)
			if err != nil {
				t.Fatal(err)
			}
			var all []string
			for rows.Next() {
				var s string
				if err := rows.Scan(&s); err != nil {
					t.Fatal(err)
				}
				all = append(all, s)
			}
			rows.Close()
			for _, seat := range all {
				var want sql.NullInt64
				var wantRank int64
				var opened string
				if err := db.QueryRow(`SELECT s."id", COALESCE(e."sitting", 0), COALESCE(s."opened_by", '')
				    FROM (SELECT max("id") AS "id" FROM "sittings" WHERE "seat_id" = ?) m
				    LEFT JOIN "sittings" s ON s."id" = m."id" LEFT JOIN "events" e ON e."id" = m."id"`, seat).
					Scan(&want, &wantRank, &opened); err != nil {
					t.Fatal(err)
				}
				got, rank, err := recordsql.LatestSitting(db, seat)
				if err != nil {
					t.Fatal(err)
				}
				if got != want || rank != wantRank {
					t.Errorf("%s: LatestSitting = %v rank %d; the sittings view says %v rank %d", seat, got, rank, want, wantRank)
				}
				seats++
				if opened == "sitting_open" {
					byBracket++
				}
			}
		})
	}
	if byBracket == 0 {
		t.Errorf("across every archived run (%d seats), no seat's latest sitting was opened by a bracket — the bracket half of the lookup was never compared", seats)
	}
}
