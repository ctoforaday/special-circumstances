package migrate_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// rankedWindows is the two windows ranked over the stored sittings by a read: each sitting numbered
// among its owner's, and the chair's numbered among themselves, with a row's epoch the newest chair
// sitting at or before it. It is the oracle the stored columns are held to, and it lives here only.
const rankedWindows = `
WITH "seat_sitting" AS MATERIALIZED (
  SELECT "id", row_number() OVER (PARTITION BY "seat_id" ORDER BY "id") AS "n" FROM "sittings"
),
"chair_sitting" AS MATERIALIZED (
  SELECT "id", row_number() OVER (ORDER BY "id") AS "n" FROM "sittings" WHERE "seat_id" = 'red-chair'
)
SELECT e."id", e."sitting", e."epoch",
       COALESCE(ss."n", 0),
       COALESCE((SELECT cs."n" FROM "chair_sitting" cs WHERE cs."id" <= e."id"
                  ORDER BY cs."id" DESC LIMIT 1), 0)
FROM "events" e
LEFT JOIN "seat_sitting" ss ON ss."id" = e."sitting_id"
ORDER BY e."id"`

// THE STORED WINDOWS ARE THE RANKED ONES, ON EVERY ARCHIVED RUN. A migration replays each run
// through the write path, which stamps sitting and epoch beside sitting_id; the ranking above reads
// the same sitting_id back and counts. Every row of every migrated run must agree — and some run must
// reach a second sitting and a chair epoch, or the agreement is between zeros.
func TestTheStoredWindowsAreTheRankedOnesOnEveryArchivedRun(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	var maxSitting, maxEpoch int64
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			_, runDir := migrate.MigratedArchive(t, name)
			db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rows, err := db.Query(rankedWindows)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			n, wrong := 0, 0
			for rows.Next() {
				var id, sitting, epoch, rankedSitting, rankedEpoch int64
				if err := rows.Scan(&id, &sitting, &epoch, &rankedSitting, &rankedEpoch); err != nil {
					t.Fatal(err)
				}
				n++
				if sitting != rankedSitting || epoch != rankedEpoch {
					if wrong++; wrong <= 5 {
						t.Errorf("event %d stores sitting %d, epoch %d; ranked over the record it is sitting %d, epoch %d",
							id, sitting, epoch, rankedSitting, rankedEpoch)
					}
				}
				maxSitting, maxEpoch = max(maxSitting, sitting), max(maxEpoch, epoch)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if n == 0 {
				t.Fatalf("%s migrated to a record with no events, so nothing was compared", name)
			}
			if wrong > 5 {
				t.Errorf("... and %d more of %d events disagree", wrong-5, n)
			}
		})
	}
	if maxSitting < 2 || maxEpoch < 2 {
		t.Errorf("across every archived run the largest stored sitting is %d and epoch %d — the comparison never left the first sitting or epoch", maxSitting, maxEpoch)
	}
}
