package recordsql

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// checkpointPart is a part with one sentence per anchor the tests place.
func checkpointPart(n int) string {
	var b strings.Builder
	b.WriteString("# Report\n\n")
	for k := range n {
		fmt.Fprintf(&b, "Sentence %d stands here. ", k)
	}
	return b.String() + "\n"
}

func anchorOn(k int) *recordpb.Anchor {
	return &recordpb.Anchor{Id: proto.String(fmt.Sprintf("F-%08x", 0xa0000000+k)), Location: proto.String(fmt.Sprintf("Sentence %d stands here", k))}
}

// checkpointAt is the event the one checkpoint row was stamped at, 0 where the table holds none.
func checkpointAt(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(`SELECT coalesce((SELECT "event_id" FROM "report_checkpoint"), 0)`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// stamps is the events the checkpoint was restamped at, read after each insert.
type stamps []int64

func (s *stamps) after(t *testing.T, db *sql.DB) {
	t.Helper()
	if at := checkpointAt(t, db); at != 0 && (len(*s) == 0 || (*s)[len(*s)-1] != at) {
		*s = append(*s, at)
	}
}

// THE CHECKPOINT IS RESTAMPED AT THE FIRST OP AND AT EVERY reportCheckpointEvery-TH OP AFTER IT, and
// an event that adds no op restamps nothing. The report read after every event, from the newest checkpoint and the
// ops after it, is the replay of every op from the empty text.
func TestACheckpointLandsAtTheFirstOpAndAtEverySpacingAfter(t *testing.T) {
	db := store(t)
	ord := int32(0)
	var got stamps
	insert := func(typ recordpb.EventType, body proto.Message) int64 {
		ord++
		id, err := Insert(db, event(t, ord, typ, body))
		if err != nil {
			t.Fatal(err)
		}
		got.after(t, db)
		ops, err := ReportOpsBetween(db, 0, Newest)
		if err != nil {
			t.Fatal(err)
		}
		whole, err := reportproj.Replay("", false, ops)
		if err != nil {
			t.Fatal(err)
		}
		src, err := ReportSourceAt(db)
		if err != nil {
			t.Fatalf("after event %d: %v", id, err)
		}
		if len(src.Ops) >= reportCheckpointEvery {
			t.Errorf("after event %d the read replays %d ops", id, len(src.Ops))
		}
		if got, err := src.Fold(reportproj.Replay); err != nil || got != whole {
			t.Errorf("after event %d the read gives (%q, %v), want the whole replay %q", id, got, err, whole)
		}
		return id
	}
	part := insert(recordpb.EventType_EVENT_TYPE_REPORT_PART, &recordpb.ReportPart{Text: proto.String(checkpointPart(20))})
	var anchors []int64
	for k := range 20 {
		anchors = append(anchors, insert(recordpb.EventType_EVENT_TYPE_ANCHOR, anchorOn(k)))
		insert(recordpb.EventType_EVENT_TYPE_LOG, &recordpb.Log{Text: proto.String("no op"), Type: recordpb.LogType_LOG_TYPE_FRICTION.Enum()})
	}
	want := []int64{part, anchors[reportCheckpointEvery-1], anchors[2*reportCheckpointEvery-1]}
	if !slices.Equal(got, want) {
		t.Fatalf("checkpoint restamped at events %v, want %v", got, want)
	}
}

// rawInsert writes an event and its body row past InsertTx, as no write path does: the record holds
// ops the stamp never saw.
func rawInsert(t *testing.T, db *sql.DB, id int64, typ, table string, cols string, vals ...any) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO "events" ("id", "seat_id", "ts", "type", "sitting", "epoch") VALUES (?, 'red-chair', '2026-01-01T00:00:00Z', ?, 0, 0)`, id, typ); err != nil {
		t.Fatal(err)
	}
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(vals)), ", ")
	if _, err := db.Exec(fmt.Sprintf(`INSERT INTO %q ("event_id", %s) VALUES (?, %s)`, table, cols, marks), append([]any{id}, vals...)...); err != nil {
		t.Fatal(err)
	}
}

// A MISSING CHECKPOINT IS AN ERROR, NEVER A LONGER REPLAY: ops with no checkpoint below them, and a
// tail as long as the spacing, are each a record this write path did not write, and the read says so.
func TestAMissingCheckpointIsAnError(t *testing.T) {
	t.Run("ops below no checkpoint", func(t *testing.T) {
		db := store(t)
		rawInsert(t, db, 1, "report_part", "report_part", `"text"`, checkpointPart(1))
		if _, err := ReportSourceAt(db); err == nil || !strings.Contains(err.Error(), "no report checkpoint at or below them") {
			t.Fatalf("the read = %v, want the missing checkpoint named", err)
		}
	})
	t.Run("a tail as long as the spacing", func(t *testing.T) {
		db := store(t)
		if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_REPORT_PART, &recordpb.ReportPart{Text: proto.String(checkpointPart(reportCheckpointEvery))})); err != nil {
			t.Fatal(err)
		}
		for k := range reportCheckpointEvery {
			if k == reportCheckpointEvery-1 {
				if _, err := ReportSourceAt(db); err != nil {
					t.Fatalf("a read before the tail reaches the spacing: %v", err)
				}
			}
			a := anchorOn(k)
			rawInsert(t, db, int64(2+k), "anchor", "anchor", `"id", "location"`, a.GetId(), a.GetLocation())
		}
		if _, err := ReportSourceAt(db); err == nil || !strings.Contains(err.Error(), "after its checkpoint at event 1 and none since") {
			t.Fatalf("the read = %v, want the missing checkpoint named", err)
		}
	})
}

// AN OP THE REPORT CANNOT REPLAY THROUGH LANDS, AND ITS CHECKPOINT HOLDS THE FAILURE: a read after it
// or any later event states the replay's own words, as the replay from the empty text does, and a
// later checkpoint carries the failure on.
func TestAnOpTheReportCannotReplayThroughIsStoredAsTheFailure(t *testing.T) {
	db := store(t)
	var got stamps
	for k := range reportCheckpointEvery + 1 {
		if _, err := Insert(db, event(t, int32(1+k), recordpb.EventType_EVENT_TYPE_ANCHOR, anchorOn(k))); err != nil {
			t.Fatalf("an anchor before any part was refused at the insert: %v", err)
		}
		got.after(t, db)
		var failure sql.NullString
		if err := db.QueryRow(`SELECT "failure" FROM "report_checkpoint" WHERE "text" IS NULL`).Scan(&failure); err != nil || failure.String != reportproj.NoPart {
			t.Fatalf("after event %d the checkpoint holds failure %q (%v), want %q", 1+k, failure.String, err, reportproj.NoPart)
		}
		src, err := ReportSourceAt(db)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := src.Fold(reportproj.Replay); err == nil || err.Error() != reportproj.NoPart {
			t.Errorf("after event %d the read = %v, want %q", 1+k, err, reportproj.NoPart)
		}
	}
	if !slices.Equal(got, []int64{1, 1 + reportCheckpointEvery}) {
		t.Fatalf("checkpoint restamped at events %v", got)
	}
}

// THE TABLE HOLDS ONE ROW, the report as it stands: a restamp replaces it, and the schema refuses a
// second.
func TestTheCheckpointIsOneRow(t *testing.T) {
	db := store(t)
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_REPORT_PART, &recordpb.ReportPart{Text: proto.String(checkpointPart(3 * reportCheckpointEvery))})); err != nil {
		t.Fatal(err)
	}
	for k := range 3 * reportCheckpointEvery {
		if _, err := Insert(db, event(t, int32(2+k), recordpb.EventType_EVENT_TYPE_ANCHOR, anchorOn(k))); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "report_checkpoint"`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("the table holds %d row(s) (%v) after three restamps, want one", n, err)
	}
	if at := checkpointAt(t, db); at != int64(1+3*reportCheckpointEvery) {
		t.Errorf("the row is stamped at event %d, want the latest restamp %d", at, 1+3*reportCheckpointEvery)
	}
	for _, stmt := range []string{
		`INSERT INTO "report_checkpoint" ("one", "event_id", "text") VALUES (2, 1, 'x')`,
		`INSERT INTO "report_checkpoint" ("one", "event_id", "text") VALUES (1, 1, 'x')`,
	} {
		if _, err := db.Exec(stmt); err == nil {
			t.Errorf("%s was accepted, want a second row refused", stmt)
		}
	}
}
