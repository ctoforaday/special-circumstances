package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE REPORT'S REPLAY IS A FOLD. A write folds the ops another write added onto the render it made
// before taking the lock, and reads that as the report it lands on (record.RegisterReportRenderer), so
// replaying ops[k:] over what ops[:k] replayed to must give the whole replay's bytes, at every k. Asked
// of archived runs, whose ops are real edits, placements and retires rather than a fixture's.
func TestTheReportReplayIsAFold(t *testing.T) {
	for _, name := range []string{"2026-10-12_is-91-prime-wave-b.tar.gz", "2026-10-13_is-91-prime-m23.tar.gz", "2026-09-11_is-91-prime-b9.tar.gz"} {
		t.Run(name, func(t *testing.T) {
			_, dst := migrateArchive(t, name)
			ops, err := record.ReportProjection(dst)
			if err != nil || len(ops) == 0 || ops[0].Kind != "append" {
				t.Fatalf("the report projection: %d ops, %v", len(ops), err)
			}
			whole, err := reportproj.Replay("", false, ops)
			if err != nil {
				t.Fatalf("the archived run does not replay: %v", err)
			}
			if len(ops) < 20 {
				t.Fatalf("%d ops is too few to say anything about a fold", len(ops))
			}
			for k := range len(ops) + 1 {
				if k == 0 {
					continue // the empty fold is no report: NoPart, which the render states
				}
				prefix, err := reportproj.Replay("", false, ops[:k])
				if err != nil {
					t.Fatalf("the first %d of %d ops do not replay: %v", k, len(ops), err)
				}
				folded, err := reportproj.Replay(prefix, true, ops[k:])
				if err != nil {
					t.Fatalf("ops %d.. over the first %d's text: %v", k, k, err)
				}
				if folded != whole {
					t.Fatalf("ops %d.. folded over the first %d's text differ from the whole replay", k, k)
				}
			}
		})
	}
}
