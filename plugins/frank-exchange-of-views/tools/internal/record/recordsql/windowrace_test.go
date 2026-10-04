package recordsql_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A LOAD READS EACH EVENT'S SITTING AND THAT SITTING'S OWNER AS ONE SNAPSHOT.
//
// Events reads every row's stored sitting_id and the owner of that sitting. Read as two statements
// on a pooled handle, a sitting opened between them loads with its id and no owner: the seat's
// latest sitting is then the one before, and every reader of "this sitting" — the work list, the
// repair claim — answers about the wrong one. A writer opening sittings while loads run must never
// leave a loaded event in a sitting whose owner the same load does not know.
func TestALoadNeverSeesASittingWithoutItsOwner(t *testing.T) {
	dir := t.TempDir()
	recordtest.Seed(t, dir, recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}))
	db, err := recordsql.Open(runtest.Open(t, dir).Dir() + "/records/record.db")
	if err != nil {
		t.Fatal(err)
	}

	const opens, readers = 200, 4
	done := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, readers+1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)
		for i := 1; i <= opens; i++ {
			seat := fmt.Sprintf("red-lens-%d", i%5)
			ev := recordtest.At(t, seat, fmt.Sprintf("%s:register:#%d", seat, i), &recordpb.Register{})
			if _, err := recordsql.Insert(db, ev); err != nil {
				errs <- err
				return
			}
		}
	}()
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				evs, ws, err := recordsql.Events(db)
				if err != nil {
					errs <- err
					return
				}
				owner := map[int64]string{}
				for i, w := range ws {
					if w.SittingID == 0 {
						continue
					}
					if w.Owner == "" {
						errs <- fmt.Errorf("event %d (%s by %s) loaded in sitting %d with no owner", w.ID, evs[i].GetType(), evs[i].GetSeatId(), w.SittingID)
						return
					}
					if o, seen := owner[w.SittingID]; seen && o != w.Owner {
						errs <- fmt.Errorf("sitting %d loaded with two owners in one load: %s and %s", w.SittingID, o, w.Owner)
						return
					}
					owner[w.SittingID] = w.Owner
					if seat, opens := w.Opens(); opens && seat != evs[i].GetSeatId() {
						errs <- fmt.Errorf("register %d by %s loaded as opening %s's sitting", w.ID, evs[i].GetSeatId(), seat)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
