package migrate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EVERY ARCHIVED RUN MIGRATES, with nothing refused. The per-run tests each pick one or two
// tarballs, so a write-path check added later can refuse a run none of them covers, and every act
// citing a refused act cascades after it while the suite stays green. This is the census, not a
// sample.
//
// EVERY SEAT A MIGRATED RECORD HOLDS IS A SEAT. An archive spells its seats with rounds, lens
// numbers and four bench ids; remap's `seat` translates each, and the role of an event's seat is
// read from the roster with no second spelling admitted. An id the translation leaves off the
// roster has no role, so every act under it renders as no party's — this is what holds the
// translation to the whole archive rather than to the ids remap_test.go lists.
func TestEveryArchivedRunMigrates(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			res, run := migrateArchive(t, name)
			if len(res.Refusals) != 0 {
				t.Errorf("%d event(s) refused migrating %s; the first: %+v", len(res.Refusals), name, res.Refusals[0])
			}
			merged, err := record.MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			// The harness writes the cast and the sitting spans under its own id (record.HarnessSeat):
			// it is no party, and no role is the answer for it.
			offRoster, seated := map[string]int{}, 0
			for _, e := range merged.Events {
				switch id := e.GetSeatId(); {
				case id == record.HarnessSeat:
					if got := record.PartyOf(e); got != "" {
						t.Errorf("%s: the harness's own events are attributed to the %s party", name, got)
					}
				case seatclass.RoleOfSeat(id) == "" || record.PartyOf(e) != seatclass.RoleOfSeat(id):
					offRoster[id]++
				default:
					seated++
				}
			}
			for id, n := range offRoster {
				t.Errorf("%s: %d migrated event(s) are attributed to %q, which is no seat on the roster", name, n, id)
			}
			if seated == 0 {
				t.Errorf("%s: no migrated event is attributed to a seat, so no seat was checked", name)
			}
		})
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}
