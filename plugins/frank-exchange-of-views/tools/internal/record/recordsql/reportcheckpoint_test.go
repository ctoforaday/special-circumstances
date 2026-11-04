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

func checkpointEvents(t *testing.T, db *sql.DB) []int64 {
	t.Helper()
	rows, err := db.Query(`SELECT "event_id" FROM "report_checkpoint" ORDER BY "event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		out = append(out, id)
	}
	return out
}

// A CHECKPOINT LANDS AT THE FIRST OP AND AT EVERY reportCheckpointEvery-TH OP AFTER IT, and an event
// that adds no op lands none. The report as of every event, read from the checkpoint at or below it
// and the ops after it, is the replay of every op up to it from the empty text.
func TestACheckpointLandsAtTheFirstOpAndAtEverySpacingAfter(t *testing.T) {
	db := store(t)
	ord := int32(0)
	insert := func(typ recordpb.EventType, body proto.Message) int64 {
		ord++
		id, err := Insert(db, event(t, ord, typ, body))
		if err != nil {
			t.Fatal(err)
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
	if got := checkpointEvents(t, db); !slices.Equal(got, want) {
		t.Fatalf("checkpoints at events %v, want %v", got, want)
	}
	for _, at := range append([]int64{part}, anchors...) {
		ops, err := ReportOpsBetween(db, 0, at)
		if err != nil {
			t.Fatal(err)
		}
		whole, err := reportproj.Replay("", false, ops)
		if err != nil {
			t.Fatal(err)
		}
		src, err := ReportSourceAt(db, at)
		if err != nil {
			t.Fatalf("as of event %d: %v", at, err)
		}
		if len(src.Ops) >= reportCheckpointEvery {
			t.Errorf("as of event %d the read replays %d ops", at, len(src.Ops))
		}
		if got, err := src.Fold(reportproj.Replay); err != nil || got != whole {
			t.Errorf("as of event %d the read gives (%q, %v), want the whole replay %q", at, got, err, whole)
		}
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
		if _, err := ReportSourceAt(db, Newest); err == nil || !strings.Contains(err.Error(), "no report checkpoint at or below them") {
			t.Fatalf("the read = %v, want the missing checkpoint named", err)
		}
	})
	t.Run("a tail as long as the spacing", func(t *testing.T) {
		db := store(t)
		if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_REPORT_PART, &recordpb.ReportPart{Text: proto.String(checkpointPart(reportCheckpointEvery))})); err != nil {
			t.Fatal(err)
		}
		for k := range reportCheckpointEvery {
			a := anchorOn(k)
			rawInsert(t, db, int64(2+k), "anchor", "anchor", `"id", "location"`, a.GetId(), a.GetLocation())
		}
		if _, err := ReportSourceAt(db, Newest); err == nil || !strings.Contains(err.Error(), "after its checkpoint at event 1 and none since") {
			t.Fatalf("the read = %v, want the missing checkpoint named", err)
		}
		if _, err := ReportSourceAt(db, int64(reportCheckpointEvery)); err != nil {
			t.Fatalf("a read as of an event before the tail reaches the spacing: %v", err)
		}
	})
}

// AN OP THE REPORT CANNOT REPLAY THROUGH LANDS, AND ITS CHECKPOINT HOLDS THE FAILURE: a read as of it
// or any later event states the replay's own words, as the replay from the empty text does, and a
// later checkpoint carries the failure on.
func TestAnOpTheReportCannotReplayThroughIsStoredAsTheFailure(t *testing.T) {
	db := store(t)
	for k := range reportCheckpointEvery + 1 {
		if _, err := Insert(db, event(t, int32(1+k), recordpb.EventType_EVENT_TYPE_ANCHOR, anchorOn(k))); err != nil {
			t.Fatalf("an anchor before any part was refused at the insert: %v", err)
		}
	}
	if got := checkpointEvents(t, db); !slices.Equal(got, []int64{1, 1 + reportCheckpointEvery}) {
		t.Fatalf("checkpoints at events %v", got)
	}
	var failures int
	if err := db.QueryRow(`SELECT count(*) FROM "report_checkpoint" WHERE "failure" = ? AND "text" IS NULL`, reportproj.NoPart).Scan(&failures); err != nil || failures != 2 {
		t.Fatalf("%d checkpoint(s) hold the replay's failure (%v), want both", failures, err)
	}
	for _, at := range []int64{1, 4, Newest} {
		src, err := ReportSourceAt(db, at)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := src.Fold(reportproj.Replay); err == nil || err.Error() != reportproj.NoPart {
			t.Errorf("as of event %d the read = %v, want %q", at, err, reportproj.NoPart)
		}
	}
}

// A CHECKPOINT IS THE TEXT AN EVENT LEFT, so the record refuses changing or removing one.
func TestACheckpointIsAppendOnly(t *testing.T) {
	db := store(t)
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_REPORT_PART, &recordpb.ReportPart{Text: proto.String("# R\n")})); err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{`UPDATE "report_checkpoint" SET "text" = 'x'`, `DELETE FROM "report_checkpoint"`} {
		if _, err := db.Exec(stmt); err == nil || !strings.Contains(err.Error(), "a report checkpoint is the text an event left") {
			t.Errorf("%s = %v, want the trigger's refusal", stmt, err)
		}
	}
}
