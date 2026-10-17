package migrate_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// AN ARCHIVED RUN STORES NO PLAN, AND ITS ROWS NO REMAND EXCHANGE. Every archive predates the stored
// plan (epoch 30): migrate reconstructs none — today's fold over the record's prefix is not what the
// binary that ran it printed — so the relay audit on a migrated run reports that it compared nothing.
// The census reads every archive and fails on a stored plan or on an archive set holding no dispatch
// row, because the translation would then be tested by nothing.
func TestAnArchivedRunStoresNoPlan(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	rows := 0
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
		if err != nil || len(stored) != 0 {
			t.Errorf("%s: migrated with %d stored plan(s), %v — an archive stores none", name, len(stored), err)
		}
		for _, e := range fam.Events {
			if _, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok {
				rows++
			}
		}
	}
	if rows == 0 {
		t.Fatal("no archived run holds a dispatch row — the translation is exercised by nothing")
	}
	t.Logf("%d archived dispatch row(s) migrated; no archive stores a plan", rows)
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
