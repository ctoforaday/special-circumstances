package migrate_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A BENCH ROW WRITTEN BEFORE EPOCH 19 CONVENED THE BENCH FOR ITS DOCKET, and migrated it says so.
// The chair readied the bench only on a docketed gap then — a petition was heard off an envelope,
// with no row — so every archived bench row carries gaps, and the write now refuses a bench row
// naming no occasion. The census reads every archive: a bench row migrated with any other occasion,
// or none, fails; and an archive set holding no bench row at all fails too, because the
// translation would then be tested by nothing.
func TestAnArchivedBenchDispatchMigratesAsADocket(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	bench := 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		res, dst := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		fam, err := record.FamilyOf(dst)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
			if !ok {
				continue
			}
			if !record.SeatOwesOccasion(d.GetSeatId()) {
				if len(d.GetOccasions()) != 0 {
					t.Errorf("%s: %s's row migrated with occasions %v — only the bench's row carries one", name, d.GetSeatId(), d.GetOccasions())
				}
				continue
			}
			bench++
			if !slices.Equal(d.GetOccasions(), []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}) || len(d.GetGapIds()) == 0 {
				t.Errorf("%s: a bench row migrated as %v on %v, want [docket] on its gaps", name, d.GetOccasions(), d.GetGapIds())
			}
		}
	}
	if bench == 0 {
		t.Fatal("no archived run holds a bench dispatch row — the translation is exercised by nothing")
	}
	t.Logf("%d archived bench row(s) migrated as a docket", bench)
}

// AN ARCHIVED RUN'S STORED PLANS ARE THE ONES IT STORED, AND ITS ROWS NO REMAND EXCHANGE. An archive
// written before the stored plan (epoch 31) migrates with none: migrate reconstructs none — today's
// fold over the record's prefix is not what the binary that ran it printed — so the relay audit on
// such a run reports that it compared nothing. An archive written at this binary's epoch carries the
// plans it stored, byte for byte. The census reads every archive and fails on a plan the source did
// not store, a stored plan lost or altered, or an archive set holding no dispatch row, because the
// translation would then be tested by nothing.
func TestAnArchivedRunStoresOnlyThePlansItStored(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	rows, carried := 0, 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		res, dst := migrateArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
		}
		fam, err := record.FamilyOf(dst)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := record.StoredPlans(fam.Events)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		// THE SOURCE'S OWN PLANS, read as the source holds them — the old-record reader opens a
		// record of any epoch; one older than 31 stores none.
		var want []string
		if res.SourceEpoch != nil && *res.SourceEpoch >= 31 {
			src, err := migrate.OpenSQLite(filepath.Join(recordtest.ExtractArchive(t, name), "records"), t.TempDir())
			if err != nil {
				t.Fatalf("%s: the source does not open: %v", name, err)
			}
			evs, err := src.Events()
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			for _, ev := range evs {
				if ev.Word == "plan" {
					printed, _ := ev.Fields["printed"].(string)
					want = append(want, printed)
				}
			}
			if err := src.Close(); err != nil {
				t.Fatal(err)
			}
		}
		var got []string
		for _, p := range stored {
			got = append(got, p.Printed)
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s: migrated with %d stored plan(s), the source stored %d — a migrated record stores exactly the source's plans, byte for byte", name, len(got), len(want))
		}
		carried += len(want)
		for _, e := range fam.Events {
			if _, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok {
				rows++
			}
		}
	}
	if rows == 0 {
		t.Fatal("no archived run holds a dispatch row — the translation is exercised by nothing")
	}
	t.Logf("%d archived dispatch row(s) migrated; %d stored plan(s) carried", rows, carried)
}

// A ROW'S REMAND EXCHANGE IS NOT CARRIED: the row is translated word for word without it.
func TestADispatchRowsRemandOwedIsNotCarried(t *testing.T) {
	bodies, err := migrate.Entries().Translate(migrate.OldEvent{ID: 9, Word: "dispatch",
		Fields: map[string]any{"pin": int64(4), "seat_id": "blue-respond"},
		Lists:  map[string][]string{"gap_ids": {"G-00000001"}, "remand_owed": {"G-00000001"}}})
	if err != nil || len(bodies) != 1 {
		t.Fatalf("translate = %d bodies, %v", len(bodies), err)
	}
	d, ok := bodies[0].(*recordpb.Dispatch)
	if !ok || d.GetPin() != 4 || d.GetSeatId() != "blue-respond" || !slices.Equal(d.GetGapIds(), []string{"G-00000001"}) {
		t.Fatalf("translated to %+v", bodies[0])
	}
}
