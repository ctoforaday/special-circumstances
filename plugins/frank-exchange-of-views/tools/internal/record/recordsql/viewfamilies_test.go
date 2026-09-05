package recordsql

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
)

// THE FAMILY VIEWS ARE TESTED IN SQL, INCLUDING THE STATES THE WRITE GUARDS FORBID. Insert
// bypasses record.Append's validation on purpose: a legacy record can carry two rulings on one
// motion, and the whole reason motion_answers states first-wins ONCE is that such a record must
// read as one answered motion, not multiply every join. A view tested only through the guarded
// write path is tested only on the states the guards permit — which is the half that was never
// in danger.

func mintRow(t *testing.T, db *sql.DB, ord int32, id string, kind recordpb.CheckKind, supersedes ...string) {
	t.Helper()
	m := &recordpb.Mint{
		GapId:           proto.String(id),
		Class:           proto.String("scope-creep"),
		Problem:         proto.String("p"),
		AcceptanceCheck: proto.String("a"),
		CheckKind:       kind.Enum(),
		Likelihood:      recordpb.Grade_GRADE_MEDIUM.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
		Supersedes:      supersedes,
	}
	if _, err := Insert(db, event(t, ord, recordpb.EventType_EVENT_TYPE_MINT, m)); err != nil {
		t.Fatal(err)
	}
}

func TestTheGapViewAnswersTheGapFamilyAtOnce(t *testing.T) {
	db := store(t)
	mintRow(t, db, 0, "R1-1", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	mintRow(t, db, 1, "R1-2", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "R1-1")
	mintRow(t, db, 2, "R1-3", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	// Two regrades on R1-1: impact moves twice (the LATEST wins), severity never touched
	// (the mint's answer stands — here the mint carried none, so NULL stands).
	for i, g := range []recordpb.Grade{recordpb.Grade_GRADE_HIGH, recordpb.Grade_GRADE_LOW} {
		if _, err := Insert(db, event(t, int32(3+i), recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{
			GapId: proto.String("R1-1"), Impact: g.Enum(), Basis: proto.String("moved")})); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Insert(db, event(t, 5, recordpb.EventType_EVENT_TYPE_PROOF, &recordpb.Proof{
		Answers: proto.String("R1-3"), Text: proto.String("t"), Script: proto.String("s"), Exit: proto.Int32(0)})); err != nil {
		t.Fatal(err)
	}

	type row struct {
		curImpact, curLik       sql.NullString
		proofAnswered, awaiting bool
		supersededBy            sql.NullString
		stranded                bool
	}
	read := func(id string) row {
		t.Helper()
		var r row
		if err := db.QueryRow(`SELECT "current_impact", "current_likelihood", "proof_answered",
		    "awaiting_proof", "superseded_by", "stranded" FROM "gap" WHERE "gap_id" = ?`, id).
			Scan(&r.curImpact, &r.curLik, &r.proofAnswered, &r.awaiting, &r.supersededBy, &r.stranded); err != nil {
			t.Fatal(err)
		}
		return r
	}

	r1 := read("R1-1")
	if r1.curImpact.String != "low" {
		t.Errorf("current_impact = %q, want the LATEST regrade's word", r1.curImpact.String)
	}
	if r1.curLik.String != "medium" {
		t.Errorf("current_likelihood = %q, want the mint's grade on an axis no regrade touched", r1.curLik.String)
	}
	if r1.proofAnswered || !r1.awaiting {
		t.Errorf("R1-1 = (proof_answered=%v, awaiting=%v): an open computation gap with no proof is the debt", r1.proofAnswered, r1.awaiting)
	}
	if r1.supersededBy.String != "R1-2" || !r1.stranded {
		t.Errorf("R1-1 = (superseded_by=%q, stranded=%v): an open superseded ancestor is a broken promise", r1.supersededBy.String, r1.stranded)
	}
	r3 := read("R1-3")
	if !r3.proofAnswered || r3.awaiting {
		t.Errorf("R1-3 = (proof_answered=%v, awaiting=%v): the recorded proof discharges the debt", r3.proofAnswered, r3.awaiting)
	}
	if r2 := read("R1-2"); r2.awaiting {
		t.Error("a document-check gap can never be awaiting proof")
	}

	// Closing the stranded ancestor keeps the lineage and drops the accusation.
	if _, err := Insert(db, event(t, 6, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId: proto.String("R1-1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		Prose: proto.String("closed")})); err != nil {
		t.Fatal(err)
	}
	if r := read("R1-1"); r.stranded || r.awaiting {
		t.Errorf("after close: (stranded=%v, awaiting=%v) — a closed gap owes nothing", r.stranded, r.awaiting)
	}
}

func TestMotionAnswersStatesFirstWinsOnce(t *testing.T) {
	db := store(t)
	mintRow(t, db, 0, "R1-1", recordpb.CheckKind_CHECK_KIND_DOCUMENT)
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
		MotionId: proto.String("M-1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R1-1")}}})); err != nil {
		t.Fatal(err)
	}
	// TWO rulings — the state the write guard refuses and a legacy record can still hold.
	for i, r := range []recordpb.GradeRuling{recordpb.GradeRuling_GRADE_RULING_REJECTED, recordpb.GradeRuling_GRADE_RULING_ACCEPTED} {
		ev := event(t, int32(2+i), recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
			MotionId: proto.String("M-1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Grade{Grade: r}})
		ev.SeatId = proto.String(fmt.Sprintf("judge-r%d", i+1))
		if _, err := Insert(db, ev); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "motion_answers" WHERE "motion_id" = 'M-1'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("motion_answers has %d rows for a twice-ruled motion, want 1 — multiplicity is the defect this view exists to stop", n)
	}
	var ruling, by string
	if err := db.QueryRow(`SELECT "ruling", "ruled_by" FROM "motion_answers" WHERE "motion_id" = 'M-1'`).
		Scan(&ruling, &by); err != nil {
		t.Fatal(err)
	}
	if ruling != "rejected" || by != "judge-r1" {
		t.Errorf("(ruling, ruled_by) = (%q, %q), want the FIRST ruling — the second does not overturn it", ruling, by)
	}
	// And motion_state, which reads this view, carries ONE row too.
	if err := db.QueryRow(`SELECT count(*) FROM "motion_state" WHERE "motion_id" = 'M-1'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("motion_state has %d rows for a twice-ruled motion, want 1", n)
	}
}

func TestLineOfInquiryCarriesTheWholeLine(t *testing.T) {
	db := store(t)
	if _, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{
		AvenueId: proto.String("Q1"), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(),
		Line: proto.String("a direction")})); err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{
		AvenueId: proto.String("Q1"), Status: recordpb.AvenueStatus_AVENUE_STATUS_PURSUED.Enum(),
		SupersedesStatus: proto.String("proposed")})); err != nil {
		t.Fatal(err)
	}
	// Two direction rulings; the LATEST wins, and its unset arm is red ruling NOTHING —
	// not an invitation to read the older word.
	if _, err := Insert(db, event(t, 2, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("Q1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION.Enum(),
		Opinion: proto.String("o"),
		Ruling:  &recordpb.MotionRule_Direction{Direction: recordpb.DirectionRuling_DIRECTION_RULING_ENDORSED}})); err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 3, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("Q1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION.Enum(),
		Opinion: proto.String("reconsidered, no word")})); err != nil {
		t.Fatal(err)
	}

	var line, status string
	var ruling sql.NullString
	if err := db.QueryRow(`SELECT "line", "status", "direction_ruling" FROM "line_of_inquiry" WHERE "avenue_id" = 'Q1'`).
		Scan(&line, &status, &ruling); err != nil {
		t.Fatal(err)
	}
	if line != "a direction" {
		t.Errorf("line = %q, want the PROPOSAL's substance — a move must not blank it", line)
	}
	if status != "pursued" {
		t.Errorf("status = %q, want the latest move's", status)
	}
	if ruling.Valid {
		t.Errorf("direction_ruling = %q, want NULL — the latest ruling carried no word, and that is the answer", ruling.String)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "line_of_inquiry"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("line_of_inquiry has %d rows for one line moved once, want 1", n)
	}
}
