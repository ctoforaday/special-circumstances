package migrate_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// A MIGRATION REBUILDS THE REPORT CHECKPOINT, AND A READ THROUGH IT IS THE WHOLE REPLAY. The
// migration replays every event through the write path, whose insert stamps the checkpoint, so no
// code of the migration's writes it. On every archived run, the checkpoint the record stores is
// byte for byte the replay from the empty text of every op up to its event, and so is the report as it
// stands now, read as every reader reads it — from the newest checkpoint and the ops after it. The
// whole replay is this test's own: no reader replays from the empty text.
func TestEveryArchivedRunReadsThroughItsCheckpointsAsTheWholeReplay(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	runs, checkpoints := 0, 0
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			_, dst := migrateArchive(t, name)
			ops, err := record.ReportProjection(dst)
			if err != nil {
				t.Fatal(err)
			}
			// whole(e) is the replay from the empty text of every op of an event at or below e: the text,
			// or the replay's failure. last[e] is the index past e's last op, for each event that adds one.
			type replayed struct{ text, failure string }
			last := map[int64]int{}
			var events []int64
			for k, op := range ops {
				if k+1 < len(ops) && ops[k+1].EventID == op.EventID {
					continue // the event's last op closes it
				}
				last[op.EventID] = k + 1
				events = append(events, op.EventID)
			}
			whole := func(at int64) replayed {
				text, err := reportproj.Replay("", false, ops[:last[at]])
				if err != nil {
					return replayed{failure: err.Error()}
				}
				return replayed{text: text}
			}

			db, err := sql.Open("sqlite", "file:"+filepath.Join(dst.Dir(), "records", "record.db")+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			// Every stored checkpoint is the whole replay at its event.
			rows, err := db.Query(`SELECT "event_id", "text", "failure" FROM "report_checkpoint" ORDER BY "event_id"`)
			if err != nil {
				t.Fatal(err)
			}
			stored := 0
			for rows.Next() {
				var at int64
				var text, failure sql.NullString
				if err := rows.Scan(&at, &text, &failure); err != nil {
					t.Fatal(err)
				}
				if _, ok := last[at]; !ok {
					t.Fatalf("a checkpoint at event %d, which adds no report op", at)
				}
				if want := whole(at); text.String != want.text || failure.String != want.failure {
					t.Errorf("the checkpoint at event %d is not the whole replay to it:\n  stored  %q %q\n  replay  %q %q",
						at, text.String, failure.String, want.text, want.failure)
				}
				stored++
			}
			if err := rows.Close(); err != nil {
				t.Fatal(err)
			}
			if (stored == 0) != (len(ops) == 0) {
				t.Fatalf("%d checkpoint(s) over %d report op(s)", stored, len(ops))
			}
			if len(ops) == 0 {
				return
			}

			// The report as it stands, as every reader of the run reads it.
			now, err := record.RenderRun(dst)
			failure := ""
			if err != nil {
				failure = err.Error()
			}
			if want := whole(events[len(events)-1]); now != want.text || failure != want.failure {
				t.Fatalf("the report read through the checkpoint is not the whole replay:\n  read    %q %q\n  replay  %q %q",
					now, failure, want.text, want.failure)
			}
			runs++
			checkpoints += stored
		})
	}
	// THE SWEEP SAYS WHAT IT HELD THE CHECKPOINTS TO, so an archive set that stopped holding a report
	// fails here rather than passing on nothing.
	if runs < 30 || checkpoints < 250 {
		t.Fatalf("held %d run(s) and %d checkpoint(s) to the whole replay — too few to say anything", runs, checkpoints)
	}
	t.Logf("held %d run(s) and %d checkpoint(s) to the whole replay", runs, checkpoints)
}
