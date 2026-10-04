package migrate_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
