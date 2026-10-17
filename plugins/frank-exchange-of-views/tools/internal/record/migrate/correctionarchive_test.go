package migrate

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY ARCHIVED CORRECTION CHAIN MIGRATES AS ITS HEAD, IN ITS ROOT'S PLACE. The archives hold four
// (b7 and b9 an inquiry, b8 a revision, m19 a certification). On each: the migrated record holds an
// act at the root's stamp, by the root's seat, and none at the stamp of any replacement or
// Correction; no key carries a replacement's `~`; and a chain whose head translates to no event (a
// revision) writes nothing at all. events_in counts every link and Correction, events_out the one
// act each chain is written as.
func TestEveryArchivedCorrectionChainMigratesAsItsHead(t *testing.T) {
	chains := 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			byKey := map[string]OldEvent{}
			next := map[string]string{}
			replaced := map[string]bool{}
			var corrections []OldEvent
			for _, old := range evs {
				byKey[old.Key] = old
				if old.Word == "correction" {
					c, _ := old.Fields["corrects"].(string)
					r, _ := old.Fields["replacement"].(string)
					next[c], replaced[r] = r, true
					corrections = append(corrections, old)
				}
			}
			if len(corrections) == 0 {
				return
			}
			res, dir := MigratedArchive(t, name)
			if res.In["correction"] != len(corrections) || res.Out["correction"] != 0 {
				t.Errorf("correction in %d out %d; want %d in and none out", res.In["correction"], res.Out["correction"], len(corrections))
			}
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			at := map[string][]*record.Event{} // stamp -> the migrated events written under it
			for _, e := range m.Events {
				at[e.GetTs()] = append(at[e.GetTs()], e)
				if strings.Contains(e.GetKey(), "~") {
					t.Errorf("the migrated record holds a replacement key %q", e.GetKey())
				}
			}
			for root := range next {
				if replaced[root] {
					continue
				}
				chains++
				head := root
				for r, ok := next[head]; ok; r, ok = next[head] {
					head = r
					if len(at[byKey[head].TS]) != 0 {
						t.Errorf("%s: the replacement %s's stamp %s holds %d migrated event(s), want none", name, head, byKey[head].TS, len(at[byKey[head].TS]))
					}
				}
				written := at[byKey[root].TS]
				if byKey[head].Word == "revision" {
					if len(written) != 0 {
						t.Errorf("%s: the revision chain from %s wrote %d event(s), want none", name, root, len(written))
					}
					continue
				}
				if len(written) == 0 || written[0].GetSeatId() != byKey[root].SeatID {
					t.Errorf("%s: no act by %s at the root %s's stamp %s — the chain's head is written in its root's place", name, byKey[root].SeatID, root, byKey[root].TS)
				}
			}
			for _, c := range corrections {
				if len(at[c.TS]) != 0 {
					t.Errorf("%s: the Correction %s's stamp holds %d migrated event(s), want none", name, c.Key, len(at[c.TS]))
				}
			}
		})
	}
	if chains < 4 {
		t.Fatalf("the archives held %d correction chain(s); want the four b7, b8, b9 and m19 hold", chains)
	}
}
