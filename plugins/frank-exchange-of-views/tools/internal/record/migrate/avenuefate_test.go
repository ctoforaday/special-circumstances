package migrate_test

import (
	"database/sql"
	"maps"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED FATE ARRIVES AS THE SEAT RECORDED IT. `rejected` — an approach that ran and refuted
// its hypothesis — is a fate this binary's epoch adds, and `abandoned` now means only a line that
// was started and gave no answer. An archived `abandoned` may have been a refutation: the record
// holds the seat's word and nothing that tells the two apart, so migration carries every avenue act
// forward under the fate it was filed with and re-reads none of them. The archived m20 run holds an
// abandoned avenue beside concluded and declined ones, at the epoch before the fate existed.
func TestAnArchivedAvenueFateMigratesAsRecorded(t *testing.T) {
	const archive = "2026-10-11_is-91-prime-m20.tar.gz"
	src := recordtest.RequireArchive(t, archive)
	epoch, stated := statedEpoch(t, src)
	if !stated || epoch >= record.EventSchema {
		t.Fatalf("%s states epoch %d (stated %v): this test needs a record older than the binary's %d", archive, epoch, stated, record.EventSchema)
	}
	fates := func(db *sql.DB) map[string]int {
		t.Helper()
		rows, err := db.Query(`SELECT "status", count(*) FROM "avenue" GROUP BY "status"`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		out := map[string]int{}
		for rows.Next() {
			var status string
			var n int
			if err := rows.Scan(&status, &n); err != nil {
				t.Fatal(err)
			}
			out[status] = n
		}
		return out
	}
	want := fates(openRO(t, src))
	if want["abandoned"] == 0 {
		t.Fatalf("%s holds no abandoned avenue act, so nothing here could have been re-read: %v", archive, want)
	}

	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if got, _ := statedEpoch(t, dst.Dir()); got != record.EventSchema {
		t.Errorf("the migrated record states epoch %d, this binary writes %d", got, record.EventSchema)
	}
	got := fates(openRO(t, dst.Dir()))
	if !maps.Equal(got, want) {
		t.Errorf("avenue acts by fate: the archive holds %v and the migrated record %v", want, got)
	}
	if got["rejected"] != 0 {
		t.Errorf("the migrated record holds %d rejected avenue act(s): no archived seat filed that fate", got["rejected"])
	}
	// The fold reads the same fates the acts carry: the avenue the archive left abandoned is abandoned.
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	abandoned := 0
	for _, a := range record.AvenuesOf(fam.Events, fam.At) {
		if a.Status == "rejected" {
			t.Errorf("avenue %s reads rejected after migration", a.ID)
		}
		if a.Status == "abandoned" {
			abandoned++
		}
	}
	if abandoned == 0 {
		t.Error("no avenue of the migrated record reads abandoned; the archive ships one under that fate")
	}
}
