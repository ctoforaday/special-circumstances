package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// validateRun is WouldRefuse under the name these tests use.
func validateRun(run Run, seatID string, typ recordpb.EventType, body proto.Message) error {
	return WouldRefuse(run, seatID, typ, body)
}

// recordOf is run's record for a test asking a reader that takes a Querier: the run's handle, or a
// nil one for a run that has recorded nothing.
func recordOf(t *testing.T, run Run) recordsql.Querier {
	t.Helper()
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// chairSittingRefusal is `dispatch next`'s chair-sitting refusal over run's stream as it stands.
func chairSittingRefusal(t *testing.T, run Run) error {
	t.Helper()
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	return requireChairSittingOpened(evs, win)
}

// dispatchStandsOn is whether plan is the standing dispatch of run's stream as it stands.
func dispatchStandsOn(t *testing.T, run Run, plan Plan) (bool, error) {
	t.Helper()
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		return false, err
	}
	return dispatchStands(evs, win, plan), nil
}

// viewOf is a writeView over run's handle and files, for a test asking one check outside a write.
func viewOf(t *testing.T, run Run) writeView {
	t.Helper()
	return writeView{q: recordOf(t, run), files: filesOf(run)}
}
