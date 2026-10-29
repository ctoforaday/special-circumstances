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
// fate owns, and is counted there.
func TestEveryArchivedRemandMigratesWithItsDirection(t *testing.T) {
	remands := 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			wantRemand, wantReopens := 0, 0
			for _, old := range evs {
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
			_, dir := MigratedArchive(t, name)
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
				t.Errorf("migrated %d remands with a direction and %d closing rulings with a reopen condition; the source states %d and %d",
					gotRemand, gotReopens, wantRemand, wantReopens)
			}
			remands += gotRemand
		})
	}
	if remands == 0 {
		t.Fatal("no archived run holds a remand, so no direction was carried")
	}
}
