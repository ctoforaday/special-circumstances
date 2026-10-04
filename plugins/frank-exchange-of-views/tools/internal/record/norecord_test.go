package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A RUN WITH NO RECORD IS ASKED THROUGH A NIL HANDLE, AND EVERY QUESTION ANSWERS ITS HONEST ZERO.
// openRunForRead answers a nil *sql.DB for a run with no record.db; handed on as a Querier that nil
// is not == nil, so a reader checking `q == nil` took it for a record and its first query panicked.
// Each reader that stands in for "no record yet" is asked through that handle here.
func TestEveryReaderAnswersARunWithNoRecordThroughItsNilHandle(t *testing.T) {
	run := mustRun(t, newRun(t))
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	if db != nil {
		t.Fatalf("the run has a record.db — this test needs a run with none")
	}
	var q recordsql.Querier = db
	if q == nil {
		t.Fatal("the nil handle compared equal to nil inside the interface — this test no longer reaches the case it is for")
	}
	if !noRecord(q) {
		t.Fatal("noRecord read the nil handle as a record")
	}
	if evs, _, err := eventsAt(q); err != nil || evs != nil {
		t.Errorf("eventsAt: %d events, %v", len(evs), err)
	}
	if evs, _, err := eventsOfAt(q); err != nil || evs != nil {
		t.Errorf("eventsOfAt: %d events, %v", len(evs), err)
	}
	if cast, err := castAt(q); err != nil || cast != nil {
		t.Errorf("castAt: %v, %v", cast, err)
	}
	var n int
	if found, err := queryRowAt(q, []any{&n}, `SELECT 1`); err != nil || found {
		t.Errorf("queryRowAt: found %v, %v", found, err)
	}
	if has, err := recordHasAt(q, `SELECT 1`); err != nil || has {
		t.Errorf("recordHasAt: %v, %v", has, err)
	}
	if epochs, err := epochsAt(q); err != nil || epochs != nil {
		t.Errorf("epochsAt: %v, %v", epochs, err)
	}
	if _, have, ops, err := ReportProjectionAt(q); err != nil || have || ops != nil {
		t.Errorf("ReportProjectionAt: base %v, %d ops, %v", have, len(ops), err)
	}
	if v, err := recordedOutcomeAt(q); err != nil || v != "" {
		t.Errorf("recordedOutcomeAt: %q, %v", v, err)
	}
	if r, err := workReadsAt(q); err != nil || !r.absent {
		t.Errorf("workReadsAt: absent %v, %v", r.absent, err)
	}
	if r, err := planReadsAt(q); err != nil || r.cast != nil {
		t.Errorf("planReadsAt: cast %v, %v", r.cast, err)
	}
	if r, err := verdictReadsAt(q); err != nil || r.halted || r.passed || r.cast != nil || r.plan != nil {
		t.Errorf("verdictReadsAt: %+v, %v", r, err)
	}
	if b, err := boardJSONOfRecord(q, nil, WindowIndex{}); err != nil || len(b.Open)+len(b.Closed) != 0 {
		t.Errorf("boardJSONOfRecord: %d open, %d closed, %v", len(b.Open), len(b.Closed), err)
	}
}
