package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ARCHIVED POSITION IS GIVEN THE ID A POSITION NOW CARRIES (gblock, 2026-10-30): the bench rules a
// position on its own id, and a record written before positions had one is brought forward by
// migrate, which mints each from the archive — the same id on every migration of the same archive.
// universe-m32 holds a chair position filed at a sitting that withheld a PASS its plan permitted.
//
// THE RESIDUE, ACCEPTED: no bench heard that position when it was filed, so the migrated record holds
// it as a position motion standing unruled, and the run's archived outcome stands beside it.
func TestAnArchivedPositionIsGivenItsID(t *testing.T) {
	ids := func() ([]string, []*record.Motion) {
		_, run := migrateArchive(t, "2026-10-24_is-91-prime-m32.tar.gz")
		fam, err := record.FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, e := range fam.Events {
			if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
				out = append(out, p.GetPositionId())
			}
		}
		var heard []*record.Motion
		for _, m := range record.MotionsOf(fam.Events, fam.At) {
			if m != nil && m.Subject == "position" {
				heard = append(heard, m)
			}
		}
		return out, heard
	}
	first, heard := ids()
	if len(first) == 0 {
		t.Fatal("the migrated record holds no position, so nothing here exercised the translation")
	}
	shape := flags.PositionID().Shape()
	seen := map[string]bool{}
	for _, id := range first {
		if !shape.MatchString(id) {
			t.Errorf("migrated position id %q is not in the position id shape %s", id, shape)
		}
		if seen[id] {
			t.Errorf("migrated position id %q is given twice", id)
		}
		seen[id] = true
	}
	again, _ := ids()
	if len(again) != len(first) {
		t.Fatalf("a second migration holds %d positions, the first %d", len(again), len(first))
	}
	for i := range first {
		if first[i] != again[i] {
			t.Errorf("position %d migrated as %s, then as %s — the archive's id is minted the same way every time", i, first[i], again[i])
		}
	}
	if len(heard) == 0 {
		t.Fatal("m32's withheld PASS reads as no position motion — the migrated record lost what the bench hears")
	}
	for _, m := range heard {
		if !seen[m.ID] {
			t.Errorf("position motion %s names no position on the migrated record", m.ID)
		}
		if m.Ruled() {
			t.Errorf("position motion %s reads as ruled %q — no bench heard it when it was filed", m.ID, m.Ruling)
		}
	}
}
