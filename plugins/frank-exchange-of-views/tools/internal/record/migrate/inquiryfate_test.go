package migrate_test

import (
	"database/sql"
	"maps"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED FATE ARRIVES AS THE SEAT RECORDED IT. `refuted` — an approach that ran and refuted
// its hypothesis — is a fate this binary's epoch adds, and `abandoned` now means only a line that
// was started and gave no answer. An archived `abandoned` may have been a refutation, and an
// archived `pursued` may be where an approach came to rest: the record holds the seat's word and
// nothing that tells them apart, so migration carries every avenue act forward under the fate it
// was filed with and re-reads none of them. Two archives, because one holds no act of every fate:
// m20 has abandoned beside concluded and declined, and b9 — a record older than the stated epoch
// — has pursued.
func TestAnArchivedInquiryFateMigratesAsRecorded(t *testing.T) {
	for archive, holds := range map[string]string{
		"2026-10-11_is-91-prime-m20.tar.gz": "abandoned",
		"2026-09-11_is-91-prime-b9.tar.gz":  "pursued",
	} {
		t.Run(holds, func(t *testing.T) { archivedFatesMigrateAsRecorded(t, archive, holds) })
	}
}

func archivedFatesMigrateAsRecorded(t *testing.T, archive, holds string) {
	src := recordtest.RequireArchive(t, archive)
	epoch, stated := statedEpoch(t, src)
	if stated && epoch >= record.EventSchema {
		t.Fatalf("%s states epoch %d: this test needs a record older than the binary's %d", archive, epoch, record.EventSchema)
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
	if want[holds] == 0 {
		t.Fatalf("%s holds no %s avenue act, so nothing here could have been re-read: %v", archive, holds, want)
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
	if got["refuted"] != 0 {
		t.Errorf("the migrated record holds %d refuted avenue act(s): no archived seat filed that fate", got["refuted"])
	}
	// The fold reads the same fates the acts carry: an avenue the archive left at the fate is at it.
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	held := 0
	for _, a := range record.InquiriesOf(fam.Events, fam.At) {
		if a.Status == "refuted" {
			t.Errorf("avenue %s reads refuted after migration", a.ID)
		}
		if a.Status == holds {
			held++
		}
	}
	if held == 0 {
		t.Errorf("no avenue of the migrated record reads %s; the archive ships one under that fate", holds)
	}
}
