package migrate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE FOLDS SHARE THE HANDLE'S DECODED STREAM AND LEAVE IT AS THE RECORD HOLDS IT, on every archived
// run. Each read of the whole record hands back the events the handle's stream holds, so a fold that
// edited one would edit every later read's. After the board, the pass blockers and each seat's work
// list have folded over the stream, it is held, event by event and window by window, to a read
// through a handle that keeps no stream.
func TestTheFoldsLeaveTheCachedStreamAsTheRecordHoldsIt(t *testing.T) {
	dir := filepath.Dir(archivePath(t, "2026-09-11_is-91-prime-b9.tar.gz"))
	tarballs, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if err != nil || len(tarballs) == 0 {
		t.Fatalf("no archived run found under %s: %v", dir, err)
	}
	for _, tb := range tarballs {
		name := filepath.Base(tb)
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			_, runDir := migrate.MigratedArchive(t, name)
			run, err := record.OpenRun(runDir)
			if err != nil {
				t.Fatal(err)
			}
			fam, err := record.FamilyOf(run)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := record.PassBlockers(run); err != nil {
				t.Fatal(err)
			}
			seats := map[string]bool{}
			for _, e := range fam.Events {
				seats[e.GetSeatId()] = true
			}
			for seat := range seats {
				role := ""
				switch {
				case seat == "red-chair":
					role = "chair"
				case strings.HasPrefix(seat, "red-lens-"):
					role = "lens"
				case strings.HasPrefix(seat, "blue-"):
					role = "blue"
				case seat == "judge":
					role = "bench"
				default:
					continue
				}
				if _, err := record.WorkOfSeat(run, role, seat); err != nil {
					t.Fatal(err)
				}
			}
			cached, err := record.MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			if len(cached.Events) == 0 {
				t.Fatalf("%s holds no events, so nothing was compared", name)
			}
			if cached.Events[0] != fam.Events[0] {
				t.Fatal("the board's read and the later read decoded the record separately: the stream was not shared, so this compares nothing")
			}
			sep, err := recordsql.OpenSeparate(filepath.Join(run.Records(), "record.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer sep.Close()
			want, wantWs, err := recordsql.Events(sep)
			if err != nil {
				t.Fatal(err)
			}
			if len(want) != len(cached.Events) {
				t.Fatalf("the shared stream holds %d events, the record %d", len(cached.Events), len(want))
			}
			wrong := 0
			for i, w := range want {
				if !proto.Equal(cached.Events[i], w) || cached.At.Of(cached.Events[i]) != wantWs[i] {
					if wrong++; wrong <= 5 {
						t.Errorf("event %d: the shared stream holds %v at %+v, the record %v at %+v",
							wantWs[i].ID, cached.Events[i], cached.At.Of(cached.Events[i]), w, wantWs[i])
					}
				}
			}
		})
	}
}
