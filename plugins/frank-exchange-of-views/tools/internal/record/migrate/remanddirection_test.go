package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY ARCHIVED REMAND MIGRATES WITH ITS DIRECTION, AND EVERY CLOSING RULING WITH WHAT REOPENS IT
// (#1253). An archive states both in `reopens_on`; the migrated record holds a remand's in
// `direction` and a closing ruling's in `reopens_on`, so the count of each on the migrated record is
// the count of the source's rulings of that fate that stated one, and no remand keeps a reopen
// condition. An archived opinion that stated neither migrates with a stated miss in the field its
// fate owns, and is counted there. A ruling migrate leaves out with its docket, an act on a gap a
// successor's mint had closed (epoch 62), is the residue supersededRulings states; a ruling the
// manifest lists as lost (no delivery in the bench's sitting) is the accepted loss.
func TestEveryArchivedRemandMigratesWithItsDirection(t *testing.T) {
	remands := 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			res, dir := MigratedArchive(t, name)
			remands += remandsCarried(t, archivedSource(t, recordtest.ExtractArchive(t, name)), dir, LostIDs(res), supersededRulings[name])
		})
	}
	// Every archived bench ruling on a motion about gaps is a lost act (no archived bench sitting
	// holds a delivery), so the translation is exercised on a source whose bench was delivered the
	// report before it ruled.
	t.Run("delivered bench", func(t *testing.T) {
		_, src, to := MigratedDeliveredBench(t)
		remands += remandsCarried(t, src, to, nil, 0)
	})
	if remands == 0 {
		t.Fatal("no source holds a remand, so no direction was carried")
	}
}

// remandsCarried holds the migrated record at dir to its source's rulings: the count of remands
// with a direction and of closing rulings with a reopen condition is the source's, less each ruling
// the manifest lists as lost and the superseded rulings left out with their docket. It returns the
// remands carried.
func remandsCarried(t *testing.T, src Source, dir string, lost map[int64]bool, superseded int) int {
	t.Helper()
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	wantRemand, wantReopens := 0, -superseded
	for _, old := range evs {
		if lost[old.ID] {
			continue
		}
		f := old.Fields
		if old.Word == "motion_rule" {
			f = old.Arms["docket"]
		} else if old.Word != "opinion" {
			continue
		}
		d, _ := f["disposition"].(string)
		r, _ := f["reopens_on"].(string)
		// An opinion that predates the reopens-on/final requirement states neither, and its
		// translation states that it does not: a stated "not stated" in the field its fate owns.
		if _, final := f["final"]; old.Word == "opinion" && r == "" && !final {
			r = "not stated"
		}
		// A source written after the remand gained its own direction states it there.
		dir, _ := f["direction"].(string)
		switch {
		case (d == "remanded" || d == "carried") && (r != "" || dir != ""):
			wantRemand++
		case r == "":
		default:
			wantReopens++
		}
	}
	m, err := record.MergedEvents(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	gotRemand, gotReopens := 0, 0
	for _, e := range m.Events {
		r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || r.GetDocket() == nil {
			continue
		}
		d := r.GetDocket()
		remand := d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED
		switch {
		case remand && d.GetDirection() == "":
			t.Errorf("remand %s migrated with no direction", e.GetKey())
		case remand && d.ReopensOn != nil:
			t.Errorf("remand %s migrated with a reopen condition %q", e.GetKey(), d.GetReopensOn())
		case remand:
			gotRemand++
		case d.Direction != nil:
			t.Errorf("closing ruling %s migrated with a direction %q", e.GetKey(), d.GetDirection())
		case d.ReopensOn != nil:
			gotReopens++
		}
	}
	if gotRemand != wantRemand || gotReopens != wantReopens {
		t.Errorf("migrated %d remands with a direction and %d closing rulings with a reopen condition; the source states %d and %d, less its lost rulings",
			gotRemand, gotReopens, wantRemand, wantReopens)
	}
	return gotRemand
}

// supersededRulings is, per archived run, the closing rulings with a reopen condition migrate leaves
// out with their docket motions, each on a gap a successor's mint had closed (actOnSupersededGap).
var supersededRulings = map[string]int{
	"2026-10-18_is-91-prime-m29.tar.gz":       1,
	"2026-10-19_is-91-prime-m30.tar.gz":       1,
	"2026-10-30_mersenne-31-prime-sum.tar.gz": 1,
}
