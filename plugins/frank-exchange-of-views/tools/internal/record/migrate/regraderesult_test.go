package migrate_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// overlaidResults is each regrade's grades by the overlay gap_grade derived before the write path
// stored them: on each axis, the latest regrade at or before this one that named the axis, else the
// mint's grade. It is the oracle the stored result grades are held to, and it lives here only.
const overlaidResults = `
SELECT a."event_id",
       a."result_severity", a."result_likelihood", a."result_impact", a."result_complexity_cost",
       COALESCE((SELECT b."severity" FROM "regrade" b WHERE b."gap_id" = a."gap_id" AND b."severity" IS NOT NULL
                  AND b."event_id" <= a."event_id" ORDER BY b."event_id" DESC LIMIT 1), m."severity"),
       COALESCE((SELECT b."likelihood" FROM "regrade" b WHERE b."gap_id" = a."gap_id" AND b."likelihood" IS NOT NULL
                  AND b."event_id" <= a."event_id" ORDER BY b."event_id" DESC LIMIT 1), m."likelihood"),
       COALESCE((SELECT b."impact" FROM "regrade" b WHERE b."gap_id" = a."gap_id" AND b."impact" IS NOT NULL
                  AND b."event_id" <= a."event_id" ORDER BY b."event_id" DESC LIMIT 1), m."impact"),
       COALESCE((SELECT b."complexity_cost" FROM "regrade" b WHERE b."gap_id" = a."gap_id" AND b."complexity_cost" IS NOT NULL
                  AND b."event_id" <= a."event_id" ORDER BY b."event_id" DESC LIMIT 1), m."complexity_cost"),
       (a."severity" IS NULL AND a."result_severity" IS NOT NULL)
         OR (a."likelihood" IS NULL AND a."result_likelihood" IS NOT NULL)
         OR (a."impact" IS NULL AND a."result_impact" IS NOT NULL)
         OR (a."complexity_cost" IS NULL AND a."result_complexity_cost" IS NOT NULL)
FROM "regrade" a
JOIN "mint" m ON m."gap_id" = a."gap_id"
ORDER BY a."event_id"`

// THE STORED RESULT GRADES ARE THE OVERLAID ONES, ON EVERY ARCHIVED RUN. A migration replays each
// run through the write path, which stamps every regrade's result from the gap's grades before it;
// the overlay above re-derives the same grades from the regrades and the mint. Every regrade of every
// migrated run must agree on all four axes — and some regrade must carry forward an axis it did not
// name, or the agreement never exercised the overlay, only the copy of the axes a regrade names.
func TestTheStoredRegradeResultsAreTheOverlaidOnesOnEveryArchivedRun(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	axes := [4]string{"severity", "likelihood", "impact", "complexity_cost"}
	var regrades, carried int
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			_, runDir := migrate.MigratedArchive(t, name)
			db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rows, err := db.Query(overlaidResults)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			wrong := 0
			for rows.Next() {
				var id int64
				var stored, overlaid [4]sql.NullString
				var carries bool
				if err := rows.Scan(&id, &stored[0], &stored[1], &stored[2], &stored[3],
					&overlaid[0], &overlaid[1], &overlaid[2], &overlaid[3], &carries); err != nil {
					t.Fatal(err)
				}
				regrades++
				if carries {
					carried++
				}
				for k := range axes {
					if stored[k] != overlaid[k] {
						if wrong++; wrong <= 5 {
							t.Errorf("regrade %d stores %s %q (set: %v); the overlay gives %q (set: %v)",
								id, axes[k], stored[k].String, stored[k].Valid, overlaid[k].String, overlaid[k].Valid)
						}
					}
				}
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if wrong > 5 {
				t.Errorf("... and %d more disagreeing axes", wrong-5)
			}
		})
	}
	if regrades == 0 || carried == 0 {
		t.Errorf("across every archived run %d regrades were compared and %d carried an axis they did not name — the overlay was never exercised", regrades, carried)
	}
	t.Logf("%d regrades compared, %d carrying an axis forward", regrades, carried)
}
