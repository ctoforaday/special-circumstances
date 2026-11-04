package record

import (
	"errors"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// replayAt is the time a test's replay hands each act it writes.
var replayAt = time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)

func mustBegin(t *testing.T, run Run) *Replay {
	t.Helper()
	b, err := BeginReplay(run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = b.Rollback() })
	return b
}

// countRowsIn counts a table's rows through the run's handle, once no replay holds it.
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

// A REFUSED ACT LEAVES NOTHING, EVEN WHEN ITS EVENTS ROW WAS WRITTEN, AND THE REPLAY TAKES THE NEXT.
// The middle act fails on the body's foreign key, after InsertTx wrote its events row; the savepoint
// takes the row back, the acts on either side land with consecutive ids, and the replay commits.
func TestAReplayUndoesARefusedActWrittenPastItsEventsRow(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	events, logs := countRowsIn(t, run, "events"), countRowsIn(t, run, "log")
	b := mustBegin(t, run)

	first, err := b.Write(blue, seatLog("first"), replayAt)
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.Write(blue, failsAfterTheEventsRow(), replayAt)
	if err == nil || errors.Is(err, ErrReplayFailed) {
		t.Fatalf("the act whose body breaks a constraint was answered %v; want its refusal, the replay still usable", err)
	}
	// The refusal is the BODY's insert, so the events row was written before it: the case the
	// savepoint exists for, not one validation stopped before any row.
	if !strings.Contains(err.Error(), "FOREIGN KEY") || strings.Contains(err.Error(), "recording the event") {
		t.Fatalf("the act was refused by %q; this test needs a refusal after the events row is written", err)
	}
	second, err := b.Write(blue, seatLog("second"), replayAt)
	if err != nil {
		t.Fatalf("the replay refused the act after a refused one: %v", err)
	}
	// Before the commit, the replay's own reads see both acts.
	var a, z int64
	if err := b.Querier().QueryRow(`SELECT (SELECT "id" FROM "events" WHERE "key" = ?), (SELECT "id" FROM "events" WHERE "key" = ?)`,
		first.GetKey(), second.GetKey()).Scan(&a, &z); err != nil {
		t.Fatalf("the replay's querier does not see its own acts: %v", err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if z != a+1 {
		t.Errorf("the acts around the refused one are events %d and %d; want consecutive ids, the refused act's row taken back", a, z)
	}
	if got := countRowsIn(t, run, "events") - events; got != 2 {
		t.Errorf("the replay committed %d events, want 2", got)
	}
	if got := countRowsIn(t, run, "log") - logs; got != 2 {
		t.Errorf("the replay committed %d log rows, want 2", got)
	}
}

// AN ACT THE SHAPE CHECK REFUSES leaves the replay usable, as a refused Append leaves the record.
func TestAReplayTakesTheActAfterOneTheShapeRefused(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	chair := sit(t, run, "red-chair")
	b := mustBegin(t, run)
	_, err := b.Write(chair, &recordpb.Position{PositionId: proto.String(NewID("position"))}, replayAt)
	mustRefuse(t, err, "requires --reason")
	if errors.Is(err, ErrReplayFailed) {
		t.Fatalf("a refused act was reported as the replay's failure: %v", err)
	}
	if _, err := b.Write(blue, seatLog("after"), replayAt); err != nil {
		t.Fatal(err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := countRowsIn(t, run, "log"); got != 1 {
		t.Errorf("the record holds %d log rows, want the one act after the refusal", got)
	}
}

// A SAVEPOINT THAT FAILS ENDS THE REPLAY. Past it the savepoints no longer bracket the acts, so
// every later act and the commit return the failure, and nothing the replay wrote is committed.
func TestASavepointFailurePoisonsTheReplay(t *testing.T) {
	for _, stmt := range []string{"SAVEPOINT " + actSavepoint, "RELEASE " + actSavepoint, "ROLLBACK TO " + actSavepoint} {
		t.Run(stmt, func(t *testing.T) {
			run := corrRun(t)
			blue := sit(t, run, "blue-respond")
			events := countRowsIn(t, run, "events")
			b := mustBegin(t, run)
			if _, err := b.Write(blue, seatLog("before"), replayAt); err != nil {
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
			if _, err := b.Write(blue, body, replayAt); !errors.Is(err, ErrReplayFailed) {
				t.Fatalf("a failed %s answered %v; want ErrReplayFailed", stmt, err)
			}
			SavepointTestHook = nil
			if _, err := b.Write(blue, seatLog("after"), replayAt); !errors.Is(err, ErrReplayFailed) {
				t.Errorf("an act after the failure answered %v; want ErrReplayFailed", err)
			}
			if err := b.Commit(); !errors.Is(err, ErrReplayFailed) {
				t.Errorf("the commit of a failed replay answered %v; want ErrReplayFailed", err)
			}
			if got := countRowsIn(t, run, "events") - events; got != 0 {
				t.Errorf("a failed replay left %d event(s) on the record, want none", got)
			}
		})
	}
}

// A REPLAY WRITES ITS OWN RUN AND NO OTHER: an act addressed elsewhere is refused, never written into
// the replay's record.
func TestAReplayRefusesAnActAddressedToAnotherRun(t *testing.T) {
	run, other := corrRun(t), corrRun(t)
	sit(t, run, "blue-respond")
	elsewhere := sit(t, other, "blue-respond")
	b := mustBegin(t, run)
	for _, id := range []Identity{elsewhere, {SeatID: "blue-respond"}} {
		_, err := b.Write(id, seatLog("misaddressed"), replayAt)
		mustRefuse(t, err, "an act lands on the run it names or nowhere")
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := countRowsIn(t, run, "log") + countRowsIn(t, other, "log"); got != 0 {
		t.Errorf("a misaddressed act left %d log row(s)", got)
	}
}

// A WRITE THAT NAMES NO RUN IS REFUSED BY NAME, before any handle is opened — the text a seat
// invoked outside a dispatch reads.
func TestAnAppendWithNoRunSaysSo(t *testing.T) {
	_, err := Append(Identity{SeatID: "blue-respond"}, seatLog("nowhere"))
	mustRefuse(t, err, "this write names no run")
}

// THE REPLAY STAMPS EACH ACT WITH THE TIME IT IS HANDED, and asks no clock: Now is never read.
func TestTheReplayStampsEachActWithTheTimeItIsHanded(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	was := Now
	t.Cleanup(func() { Now = was })
	Now = func() time.Time { t.Fatal("the replay read the clock"); return time.Time{} }
	b := mustBegin(t, run)
	first, second := replayAt, replayAt.Add(-time.Hour) // an archive's stamps need not rise with the ids
	for _, at := range []time.Time{first, second} {
		ev, err := b.Write(blue, seatLog(at.String()), at)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := ParseStamp(ev.GetTs()); err != nil || !got.Equal(at) {
			t.Errorf("the act handed %s was stamped %q (%v)", at, ev.GetTs(), err)
		}
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
}

// THE REPLAY WRITES WHAT ONLY AN ACT JUDGEMENT REFUSES: the shape is every writer's, and the act
// judgements are the live write's (judgeAct). A supports_with_bridge reading that names no kind is
// refused by Append, and written by the replay as the archive holds it.
func TestTheReplayAsksNoActJudgement(t *testing.T) {
	run := corrRun(t)
	lens := sit(t, run, "red-lens-evidence")
	body := func() *recordpb.Verify {
		return &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Claim: proto.String("c"),
			Url: proto.String("https://example.org/a"), Title: proto.String("A source"),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("t")}
	}
	if _, err := Append(lens, body()); err == nil || !strings.Contains(err.Error(), "supports_with_bridge requires --bridge") {
		t.Fatalf("live, a bridge with no kind: %v, want it refused", err)
	}
	b := mustBegin(t, run)
	if _, err := b.Write(lens, body(), replayAt); err != nil {
		t.Fatalf("the replay refused a bridge with no kind: %v", err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
}

// replayed writes body as the migration's replay writes it — the shape checked, no act judged — in
// a replay of its own, and commits it.
func replayed(t *testing.T, id Identity, body proto.Message) (*Event, error) {
	t.Helper()
	b := mustBegin(t, id.Run)
	ev, err := b.Write(id, body, Now())
	if cerr := b.Commit(); cerr != nil {
		t.Fatal(cerr)
	}
	return ev, err
}

// wouldReplay is WouldRefuse as the replay asks it: the shape alone, inside a write transaction
// rolled back.
func wouldReplay(t *testing.T, run Run, seatID string, typ recordpb.EventType, body proto.Message) error {
	t.Helper()
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		return validate(writeView{q: db, files: filesOf(run)}, seatID, typ, body)
	}
	tx, err := recordsql.BeginWrite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	return validate(writeView{q: tx, files: filesOf(run)}, seatID, typ, body)
}
