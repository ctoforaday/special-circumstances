package record

import (
	"errors"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// migrating sets Migrating for the rest of the test, as migrate's replay does around a batch.
func migrating(t *testing.T) {
	t.Helper()
	Migrating = true
	t.Cleanup(func() { Migrating = false })
}

func mustBegin(t *testing.T, run Run) *Batch {
	t.Helper()
	b, err := BeginBatch(run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Rollback() })
	return b
}

// countRowsIn counts a table's rows through the run's handle, once no batch holds it.
func countRowsIn(t *testing.T, run Run, table string) int {
	t.Helper()
	db, err := openRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "` + table + `"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// A body the schema refuses only AFTER the act's events row is written: LOG_SOURCE_UNSPECIFIED
// passes validation and fails the log table's enum foreign key at the body's insert.
func failsAfterTheEventsRow() *recordpb.Log {
	return &recordpb.Log{Text: proto.String("x"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource(0).Enum()}
}

// A REFUSED ACT LEAVES NOTHING, EVEN WHEN ITS EVENTS ROW WAS WRITTEN, AND THE BATCH TAKES THE NEXT.
// The middle act fails on the body's foreign key, after InsertTx wrote its events row; the savepoint
// takes the row back, the acts on either side land with consecutive ids, and the batch commits.
func TestABatchUndoesARefusedActWrittenPastItsEventsRow(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	events, logs := countRowsIn(t, run, "events"), countRowsIn(t, run, "log")
	migrating(t)
	b := mustBegin(t, run)

	first, err := b.Append(blue, seatLog("first"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.Append(blue, failsAfterTheEventsRow())
	if err == nil || errors.Is(err, ErrBatchFailed) {
		t.Fatalf("the act whose body breaks a constraint was answered %v; want its refusal, the batch still usable", err)
	}
	// The refusal is the BODY's insert, so the events row was written before it: the case the
	// savepoint exists for, not one validation stopped before any row.
	if !strings.Contains(err.Error(), "FOREIGN KEY") || strings.Contains(err.Error(), "recording the event") {
		t.Fatalf("the act was refused by %q; this test needs a refusal after the events row is written", err)
	}
	second, err := b.Append(blue, seatLog("second"))
	if err != nil {
		t.Fatalf("the batch refused the act after a refused one: %v", err)
	}
	// Before the commit, the batch's own reads see both acts.
	var a, z int64
	if err := b.Querier().QueryRow(`SELECT (SELECT "id" FROM "events" WHERE "key" = ?), (SELECT "id" FROM "events" WHERE "key" = ?)`,
		first.GetKey(), second.GetKey()).Scan(&a, &z); err != nil {
		t.Fatalf("the batch's querier does not see its own acts: %v", err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if z != a+1 {
		t.Errorf("the acts around the refused one are events %d and %d; want consecutive ids, the refused act's row taken back", a, z)
	}
	if got := countRowsIn(t, run, "events") - events; got != 2 {
		t.Errorf("the batch committed %d events, want 2", got)
	}
	if got := countRowsIn(t, run, "log") - logs; got != 2 {
		t.Errorf("the batch committed %d log rows, want 2", got)
	}
}

// AN ACT VALIDATION REFUSES leaves the batch usable, as a refused Append leaves the record.
func TestABatchTakesTheActAfterOneValidationRefused(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	migrating(t)
	b := mustBegin(t, run)
	_, err := b.Append(blue, &recordpb.Position{})
	mustRefuse(t, err, "requires --reason")
	if errors.Is(err, ErrBatchFailed) {
		t.Fatalf("a refused act was reported as the batch's failure: %v", err)
	}
	if _, err := b.Append(blue, seatLog("after")); err != nil {
		t.Fatal(err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := countRowsIn(t, run, "log"); got != 1 {
		t.Errorf("the record holds %d log rows, want the one act after the refusal", got)
	}
}

// A SAVEPOINT THAT FAILS ENDS THE BATCH. Past it the savepoints no longer bracket the acts, so every
// later act and the commit return the failure, and nothing the batch wrote is committed.
func TestASavepointFailurePoisonsTheBatch(t *testing.T) {
	for _, stmt := range []string{"SAVEPOINT " + actSavepoint, "RELEASE " + actSavepoint, "ROLLBACK TO " + actSavepoint} {
		t.Run(stmt, func(t *testing.T) {
			run := corrRun(t)
			blue := sit(t, run, "blue-respond")
			events := countRowsIn(t, run, "events")
			migrating(t)
			b := mustBegin(t, run)
			if _, err := b.Append(blue, seatLog("before")); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { SavepointTestHook = nil })
			SavepointTestHook = func(s string) error {
				if s == stmt {
					return errors.New("disk I/O error")
				}
				return nil
			}
			// ROLLBACK TO runs only on a refused act, so that case is driven by one.
			body := proto.Message(seatLog("poisoned"))
			if strings.HasPrefix(stmt, "ROLLBACK") {
				body = failsAfterTheEventsRow()
			}
			if _, err := b.Append(blue, body); !errors.Is(err, ErrBatchFailed) {
				t.Fatalf("a failed %s answered %v; want ErrBatchFailed", stmt, err)
			}
			SavepointTestHook = nil
			if _, err := b.Append(blue, seatLog("after")); !errors.Is(err, ErrBatchFailed) {
				t.Errorf("an act after the failure answered %v; want ErrBatchFailed", err)
			}
			if err := b.Commit(); !errors.Is(err, ErrBatchFailed) {
				t.Errorf("the commit of a failed batch answered %v; want ErrBatchFailed", err)
			}
			if got := countRowsIn(t, run, "events") - events; got != 0 {
				t.Errorf("a failed batch left %d event(s) on the record, want none", got)
			}
		})
	}
}

// ONLY A MIGRATION HOLDS A BATCH: without Migrating, BeginBatch refuses and opens nothing. An appeal
// or a correction through a batch therefore always runs under Migrating.
func TestABatchIsBegunOnlyUnderMigrating(t *testing.T) {
	run := corrRun(t)
	if b, err := BeginBatch(run); err == nil {
		_ = b.Rollback()
		t.Fatal("BeginBatch opened a batch with Migrating unset")
	} else {
		mustRefuse(t, err, "a batch is opened by migrate and by nothing else")
	}
}

// A BATCH WRITES ITS OWN RUN AND NO OTHER: an act addressed elsewhere is refused, never written into
// the batch's record.
func TestABatchRefusesAnActAddressedToAnotherRun(t *testing.T) {
	run, other := corrRun(t), corrRun(t)
	sit(t, run, "blue-respond")
	elsewhere := sit(t, other, "blue-respond")
	migrating(t)
	b := mustBegin(t, run)
	for _, id := range []Identity{elsewhere, {SeatID: "blue-respond"}} {
		_, err := b.Append(id, seatLog("misaddressed"))
		mustRefuse(t, err, "an act lands on the run it names or nowhere")
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := countRowsIn(t, run, "log") + countRowsIn(t, other, "log"); got != 0 {
		t.Errorf("a misaddressed act left %d log row(s)", got)
	}
}

// THE CORRECTION GUARDS ARE THE ACT'S, NOT Append's: through a batch, a correcting identity writes
// one replacement and refuses a second, as through Append.
func TestABatchHoldsACorrectionToOneReplacement(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	migrating(t)
	b := mustBegin(t, run)
	k, err := b.Append(blue, seatLog("first, wrong"))
	if err != nil {
		t.Fatal(err)
	}
	fix := correcting(blue, recordpb.EventType_EVENT_TYPE_LOG, k.GetKey(), "typo")
	rep, err := b.Append(fix, seatLog("first, right"))
	if err != nil {
		t.Fatal(err)
	}
	if want := k.GetKey() + "~1"; rep.GetKey() != want {
		t.Errorf("the replacement is keyed %q, want %q", rep.GetKey(), want)
	}
	_, err = b.Append(fix, seatLog("a second replacement"))
	mustRefuse(t, err, "this identity already wrote its correction of "+k.GetKey())
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := countRowsIn(t, run, "correction"); got != 1 {
		t.Errorf("the record holds %d correction(s), want 1", got)
	}
}

// A WRITE THAT NAMES NO RUN IS REFUSED BY NAME, before any handle is opened — the text a seat
// invoked outside a dispatch reads.
func TestAnAppendWithNoRunSaysSo(t *testing.T) {
	_, err := Append(Identity{SeatID: "blue-respond"}, seatLog("nowhere"))
	mustRefuse(t, err, "this write names no run")
}
