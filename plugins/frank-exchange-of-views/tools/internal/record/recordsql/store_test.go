package recordsql

import (
	"database/sql"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func store(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func event(t *testing.T, seq int32, typ recordpb.EventType, body proto.Message) *recordpb.Event {
	t.Helper()
	ev := &recordpb.Event{
		SeatId: proto.String("red-merge-r1"),
		Round:  proto.Int32(1),
		Seq:    proto.Int32(seq),
		Nonce:  proto.String("aaaaaaaa"),
		Ts:     proto.String("2026-01-01T00:00:00Z"),
	}
	got, err := recordpb.SetBody(ev, body)
	if err != nil {
		t.Fatal(err)
	}
	if got != typ {
		t.Fatalf("SetBody derived %v, want %v", got, typ)
	}
	ev.Type = &got
	return ev
}

// A MINT LANDS AS COLUMNS, NOT AS A BLOB.
//
// The whole reason to be here rather than in a file is that the fields are queryable. If a body
// arrived as an opaque value, the move would have bought transactions and ordering and forfeited
// joins, constraints and aggregates — which is the trade this test exists to refuse.
func TestABodyIsWrittenAsColumns(t *testing.T) {
	db := store(t)
	id, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Class:           proto.String("scope-creep"),
		Problem:         proto.String("an absence of findings is reported as an absence of risk"),
		AcceptanceCheck: proto.String("the claim names the search that produced it"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_HIGH.Enum(),
		Supersedes:      []string{"R0-4", "R0-9"},
	}))
	if err != nil {
		t.Fatal(err)
	}

	var class, problem, kind, likelihood string
	if err := db.QueryRow(`SELECT class, problem, check_kind, likelihood FROM mint WHERE event_id = ?`, id).
		Scan(&class, &problem, &kind, &likelihood); err != nil {
		t.Fatalf("the mint is not readable as columns: %v", err)
	}
	if class != "scope-creep" || kind != "document" || likelihood != "high" {
		t.Errorf("mint = (%q, %q, %q), want (scope-creep, document, high)", class, kind, likelihood)
	}

	// The enum went in as its WORD, so it joins to the vocabulary that explains it.
	var means string
	if err := db.QueryRow(`SELECT v.means FROM mint m JOIN enum_grade v ON v.value = m.likelihood WHERE m.event_id = ?`, id).Scan(&means); err != nil {
		t.Fatalf("a grade does not join to its meaning: %v", err)
	}
	if means == "" {
		t.Error("the grade joins to an empty meaning")
	}

	// A repeated field is rows, in order, so lineage is queryable rather than a string to split.
	rows, err := db.Query(`SELECT value FROM mint_supersedes WHERE event_id = ? ORDER BY ord`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 2 || got[0] != "R0-4" || got[1] != "R0-9" {
		t.Errorf("supersedes = %v, want [R0-4 R0-9]", got)
	}
}

// AN ABSENT FIELD IS NULL, NOT A ZERO.
//
// Every column is nullable for this reason. A mint with no severity is UNGRADED on that axis, and
// writing a zero would record a grade the seat never gave — the distinction the whole schema exists
// to keep, and the one a plain Go struct could not have expressed.
func TestAnAbsentFieldIsNull(t *testing.T) {
	db := store(t)
	id, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
		Class:           proto.String("c"),
		Problem:         proto.String("p"),
		AcceptanceCheck: proto.String("a"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_MEDIUM.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var severity *string
	if err := db.QueryRow(`SELECT severity FROM mint WHERE event_id = ?`, id).Scan(&severity); err != nil {
		t.Fatal(err)
	}
	if severity != nil {
		t.Errorf("severity = %q for a mint that never graded it — an ungraded axis reads as graded", *severity)
	}
}

// THE ORDERING PROBLEM IS GONE, and this is the assertion the whole move was for.
//
// Filing and ruling were written by different seats into different files, and replay merged them by
// timestamp — so a ruling could replay BEFORE its filing. record/motion.go records that in capitals
// and works around it with a second pass. Here the state is simply unwritable.
func TestARulingCannotPrecedeItsFiling(t *testing.T) {
	db := store(t)
	rule := event(t, 0, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_ACCEPTED},
	})
	if _, err := Insert(db, rule); err == nil {
		t.Fatal("a ruling was recorded against a motion that does not exist — this is the ordering " +
			"hazard the sharded log could not refuse, and the reason Motions() needs a second pass")
	}

	filing := event(t, 1, recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R1-1")}},
	})
	if _, err := Insert(db, filing); err != nil {
		t.Fatalf("the filing was refused: %v", err)
	}
	if _, err := Insert(db, rule); err != nil {
		t.Fatalf("the ruling was refused after its filing existed: %v", err)
	}

	// And the join it enables: the gap id lives on the filing, the verdict on the ruling, and one
	// query pairs them. That join is hand-written at eight readers in the file-backed record.
	var gap, verdict string
	if err := db.QueryRow(`
		SELECT g.gap_id, r.grade
		FROM motion_rule r
		JOIN motion m       ON m.motion_id = r.motion_id
		JOIN motion_grade g ON g.event_id  = m.event_id
		WHERE r.motion_id = 'M1'`).Scan(&gap, &verdict); err != nil {
		t.Fatalf("the filing/ruling join does not resolve: %v", err)
	}
	if gap != "R1-1" || verdict != "accepted" {
		t.Errorf("join = (%q, %q), want (R1-1, accepted)", gap, verdict)
	}
}

// A RULE THAT SPANS TWO FIELDS IS ENFORCED BY THE DATABASE, not only by the tool.
//
// "closed_with_regression requires a successor" cannot be NOT NULL — the column must stay nullable
// for every other closure class — so it lived as an `if` in validate. That left the DATABASE
// willing to store the exact state the tool refused, and for a record that is EVIDENCE the gap is
// real: anything writing SQL directly could create it, and the row would look legitimate.
func TestARegressionClosureMustNameItsSuccessor(t *testing.T) {
	db := store(t)
	mk := func(seq int32) *recordpb.Event {
		return event(t, seq, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
			GapId:        proto.String("R1-1"),
			ClosureClass: recordpb.ClosureClass_CLOSURE_CLASS_CLOSED_WITH_REGRESSION.Enum(),
			Prose:        proto.String("repaired, and the retry path broke"),
		})
	}
	if _, err := Insert(db, mk(0)); err == nil {
		t.Fatal("a regression closure with no successor was stored — the thread from the defect to " +
			"the gap now carrying it is lost, and the repair_regression denominator counts a closure " +
			"that resolved nothing")
	}

	ok := event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-2"),
		ClosureClass: recordpb.ClosureClass_CLOSURE_CLASS_CLOSED_WITH_REGRESSION.Enum(),
		Successor:    proto.String("R2-1"),
		Prose:        proto.String("repaired, and R2-1 carries the regression"),
	})
	if _, err := Insert(db, ok); err != nil {
		t.Fatalf("a regression closure that NAMES its successor was refused: %v", err)
	}

	// And an ordinary closure still needs no successor, or the rule is a wall rather than a gate.
	plain := event(t, 2, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-3"),
		ClosureClass: recordpb.ClosureClass_CLOSURE_CLASS_CLOSED.Enum(),
		Prose:        proto.String("verified at the leaf"),
	})
	if _, err := Insert(db, plain); err != nil {
		t.Fatalf("an ordinary closure was refused for having no successor: %v", err)
	}
}
