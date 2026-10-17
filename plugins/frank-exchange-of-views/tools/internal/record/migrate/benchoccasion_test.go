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

// A DISPATCH ROW WRITTEN BEFORE EPOCH 29 RECORDS NO REMAND EXCHANGE, and migrated it records none.
// The field did not exist, and what the row was owed is the plan at that point — a fold the record
// does not hold and migrate does not reconstruct. The census reads every archive: a migrated row
// carrying remand_owed fails, and an archive set holding no lens or blue row fails too, because the
// translation would then be tested by nothing. The rows that engage a gap the bench had remanded
// are logged: they are the rows whose relay, had a journal survived, would now depart on the field.
func TestAnArchivedDispatchRecordsNoRemandExchange(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	rows, afterRemand := 0, 0
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
		docketGap := map[string]string{} // a docket motion's id -> its gap
		remanded := map[string]bool{}    // the gaps the bench has remanded by this point
		for _, e := range fam.Events {
			if mo, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && mo.GetDocket() != nil {
				docketGap[mo.GetMotionId()] = mo.GetDocket().GetGapId()
			}
			if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetDocket().GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
				remanded[docketGap[r.GetMotionId()]] = true
			}
			d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
			if !ok || record.SeatOwesOccasion(d.GetSeatId()) {
				continue
			}
			rows++
			if len(d.GetRemandOwed()) != 0 {
				t.Errorf("%s: %s's row migrated with remand_owed %v — a row written before the field records none", name, d.GetSeatId(), d.GetRemandOwed())
			}
			if slices.ContainsFunc(d.GetGapIds(), func(g string) bool { return remanded[g] }) {
				afterRemand++
			}
		}
	}
	if rows == 0 {
		t.Fatal("no archived run holds a lens or blue dispatch row — the translation is exercised by nothing")
	}
	t.Logf("%d archived lens and blue row(s) migrated with no remand_owed; %d of them engage a gap the bench had remanded", rows, afterRemand)
}
