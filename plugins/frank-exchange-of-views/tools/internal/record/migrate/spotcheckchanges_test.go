package migrate_test

import (
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED SPOT-CHECK ARRIVES STAMPED (EventSchema 71). Before the stamp, a spot-check's areas
// carried nothing of what changed behind their pins, and its prose said what it liked. Migrated, the
// replay stamps each area's pin and the changes to the report after it off the migrated record: each
// pin is the head the lens was last served the report at, and the changes are every blue_edit and
// report_part between the earliest pin and the spot-check. A spot-check stands only on a delivery in
// the chair's sitting, which no archived chair sitting holds, so every archived spot-check is a lost
// act and the stamp is read on a delivered source.
func TestAnArchivedSpotCheckArrivesWithTheChangesBehindItsPins(t *testing.T) {
	// No archived chair sitting holds a delivery, so each archive's spot-check is a lost act: listed
	// in the manifest, and no spot-check on the migrated record.
	for _, archive := range []string{"2026-10-29_is-91-prime-m36.tar.gz", "2026-10-24_is-91-prime-m32.tar.gz"} {
		t.Run(archive, func(t *testing.T) {
			recordtest.RequireArchive(t, archive)
			res, dst := migrateArchive(t, archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			if got := migrate.LostByWord(res)["spot_check"]; got != res.In["spot_check"] || got == 0 {
				t.Errorf("%d of %d archived spot-check(s) listed lost; the chair's sitting holds no delivery", got, res.In["spot_check"])
			}
			if n := len(int64sOf(t, openRO(t, dst.Dir()), `SELECT "event_id" FROM "spot_check"`)); n != 0 {
				t.Errorf("the migrated record holds %d spot-check(s), all of them lost", n)
			}
		})
	}
	// The stamp is held on a source whose chair was delivered the report before it spot-checked: the
	// lens's area is stamped at the head the lens was served, and the change is blue's edit after it.
	t.Run("delivered chair", func(t *testing.T) {
		_, _, dir := migrate.MigratedDeliveredBench(t)
		db := openRO(t, dir)
		var spot int64
		if err := db.QueryRow(`SELECT "event_id" FROM "spot_check"`).Scan(&spot); err != nil {
			t.Fatal(err)
		}
		areas := wordsOf(t, db, `SELECT "value" FROM "spot_check_areas" ORDER BY "ord"`)
		pins := int64sOf(t, db, `SELECT "value" FROM "spot_check_pins" ORDER BY "ord"`)
		if len(areas) != 1 || len(pins) != 1 {
			t.Fatalf("the migrated spot-check names areas %v at pins %v, want the lens's one", areas, pins)
		}
		served := int64sOf(t, db, `SELECT r."head" FROM "report_read" r JOIN "events" e ON e."id" = r."event_id" WHERE e."seat_id" = '`+areas[0]+`' ORDER BY r."event_id"`)
		if len(served) == 0 || pins[0] != served[len(served)-1] {
			t.Errorf("area %s stamped at pin %d; the lens was served the report at %v", areas[0], pins[0], served)
		}
		want := int64sOf(t, db, `SELECT "id" FROM "events" WHERE "type" IN ('blue_edit', 'report_part') AND "id" > `+itoa(pins[0])+` AND "id" < `+itoa(spot)+` ORDER BY "id"`)
		got := int64sOf(t, db, `SELECT "value" FROM "spot_check_changes" ORDER BY "ord"`)
		if len(want) == 0 || !slices.Equal(got, want) {
			t.Errorf("stamped changes %v; the migrated record holds %v after pin %d", got, want, pins[0])
		}
	})
}

func int64sOf(t *testing.T, db *sql.DB, query string) []int64 {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
