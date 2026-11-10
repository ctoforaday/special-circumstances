package migrate_test

import (
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED SPOT-CHECK ARRIVES STAMPED (EventSchema 71). Before the stamp, a spot-check's areas
// carried nothing of what changed behind their pins, and its prose said what it liked. Migrated, the
// replay stamps each area's pin and the changes to the report after it off the migrated record: each
// pin is the one the chair's own last plan before it listed for that stale area, and the changes are
// every blue_edit and report_part between the earliest pin and the spot-check.
//
// THE RESIDUE, accepted: an archive from before the whole-report serve was recorded (#1458; m32 has
// no report_read) holds no pin the live rule can read, so every area stamps pin 0 and every change
// before the spot-check, where its own plan listed the dispatch-era pins.
func TestAnArchivedSpotCheckArrivesWithTheChangesBehindItsPins(t *testing.T) {
	for _, c := range []struct {
		archive string
		areas   int
		served  bool // the archive records the whole-report serve its pins are read from
	}{
		{archive: "2026-10-29_is-91-prime-m36.tar.gz", areas: 2, served: true},
		{archive: "2026-10-24_is-91-prime-m32.tar.gz", areas: 5},
	} {
		t.Run(c.archive, func(t *testing.T) {
			recordtest.RequireArchive(t, c.archive)
			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			db := openRO(t, dst.Dir())
			var spot int64
			if err := db.QueryRow(`SELECT "event_id" FROM "spot_check"`).Scan(&spot); err != nil {
				t.Fatal(err)
			}
			areas := wordsOf(t, db, `SELECT "value" FROM "spot_check_areas" ORDER BY "ord"`)
			if len(areas) != c.areas {
				t.Fatalf("the migrated spot-check names %d area(s), this test was written against %d", len(areas), c.areas)
			}
			pins := int64sOf(t, db, `SELECT "value" FROM "spot_check_pins" ORDER BY "ord"`)
			if len(pins) != len(areas) {
				t.Fatalf("%d area(s) carry %d pin(s)", len(areas), len(pins))
			}
			// THE PLAN'S PINS, read off the chair's last stored plan before the spot-check.
			var printed string
			if err := db.QueryRow(`SELECT "printed" FROM "plan" WHERE "event_id" < ? ORDER BY "event_id" DESC LIMIT 1`, spot).Scan(&printed); err != nil {
				t.Fatal(err)
			}
			var plan struct {
				StaleAreas []struct {
					SeatID string `json:"seat_id"`
					Pin    int64  `json:"pin"`
				} `json:"stale_areas"`
			}
			if err := json.Unmarshal([]byte(printed), &plan); err != nil {
				t.Fatal(err)
			}
			planPin := map[string]int64{}
			for _, a := range plan.StaleAreas {
				planPin[a.SeatID] = a.Pin
			}
			earliest := pins[0]
			for i, a := range areas {
				if !c.served {
					if pins[i] != 0 {
						t.Errorf("area %s stamped at pin %d in a record with no serve; want 0", a, pins[i])
					}
					continue
				}
				if p, ok := planPin[a]; !ok || p != pins[i] {
					t.Errorf("area %s stamped at pin %d; the plan before it lists it at %d (listed: %v)", a, pins[i], p, ok)
				}
				earliest = min(earliest, pins[i])
			}
			want := int64sOf(t, db, `SELECT "id" FROM "events" WHERE "type" IN ('blue_edit', 'report_part') AND "id" > `+itoa(earliest)+` AND "id" < `+itoa(spot)+` ORDER BY "id"`)
			got := int64sOf(t, db, `SELECT "value" FROM "spot_check_changes" ORDER BY "ord"`)
			if len(want) == 0 || !slices.Equal(got, want) {
				t.Errorf("stamped changes %v; the migrated record holds %v after pin %d", got, want, earliest)
			}
		})
	}
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
