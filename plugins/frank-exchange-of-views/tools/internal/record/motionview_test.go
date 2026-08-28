package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"
)

// A MOTION MUST BE READABLE BY THE SEAT THAT HAS TO ANSWER IT.
//
// An unruled motion blocks `merge verdict --as PASS`, and before this view the refusal's id was
// all a seat could ever learn: no read verb, no projection. A probed merge seat blocked here
// searched six views and three help pages, then ruled `rejected` on an argument it had not read.
func TestMotionsViewCarriesTheAskNotJustTheAnswer(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"blue-respond-r1", "red-merge-r1"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: s, Round: RoundIn(mustRun(t, runDir))(s)}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: RoundIn(mustRun(t, runDir))("red-merge-r1")}, &recordpb.Mint{GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("the check runs"), Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_COMPUTATION), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}); err != nil {
		t.Fatal(err)
	}
	basis := "the defect is presentational, so `certain` severity prices a rewrite as a data error"
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond-r1", Round: RoundIn(mustRun(t, runDir))("blue-respond-r1")}, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Basis:    proto.String(basis),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
			GapId:     proto.String("R1-1"),
			Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY),
			Proposed:  recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	b, err := BoardState(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	j := motionsJSONOf(b)
	if len(j.Motions) != 1 {
		t.Fatalf("projected %d motions, want 1", len(j.Motions))
	}
	m := j.Motions[0]
	// BASIS IS THE LOAD-BEARING FIELD. A view carrying id, subject and grades but not the
	// argument leaves the seat exactly as blind while looking like it solved the problem.
	if m.Basis != basis {
		t.Errorf("basis = %q, want the filer's own argument — the ruling exists to answer it", m.Basis)
	}
	if m.Ruled {
		t.Error("an unruled motion must report ruled=false: that is the field a blocked seat is looking for")
	}
	if j.Counts.Outstanding != 1 {
		t.Errorf("outstanding = %d, want 1 — this is the count that blocks a PASS", j.Counts.Outstanding)
	}
	if m.Fields["gap_id"] != "R1-1" || m.Fields["dimension"] != "severity" || m.Fields["proposed"] != "medium" {
		t.Errorf("subject-specific fields lost: %v", m.Fields)
	}

	// And once answered, the answer sits beside the ask rather than replacing it.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: RoundIn(mustRun(t, runDir))("red-merge-r1")}, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Opinion:  proto.String("the grades stand"),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED},
	}); err != nil {
		t.Fatal(err)
	}
	b, _ = BoardState(mustRun(t, runDir))
	j = motionsJSONOf(b)
	m = j.Motions[0]
	if !m.Ruled || m.Ruling != "rejected" || m.Opinion == "" {
		t.Fatalf("ruling not projected: %+v", m)
	}
	if m.Basis != basis {
		t.Error("the ask was dropped once answered — a reader cannot weigh a ruling without it")
	}
	if j.Counts.Outstanding != 0 {
		t.Errorf("outstanding = %d after ruling, want 0", j.Counts.Outstanding)
	}
}

// The PASS refusal must say how to unblock. Handing a seat an id with no read is what produced
// the fabricated ruling in the first place.
func TestThePassRefusalNamesTheRead(t *testing.T) {
	runDir := newRun(t)
	for _, sid := range []string{"blue-respond-r1", "red-merge-r1"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: sid, Round: RoundIn(mustRun(t, runDir))(sid)}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: RoundIn(mustRun(t, runDir))("red-merge-r1")}, &recordpb.Mint{GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("the check runs"), Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond-r1", Round: RoundIn(mustRun(t, runDir))("blue-respond-r1")}, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Basis:    proto.String("b"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
			GapId:     proto.String("R1-1"),
			Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY),
			Proposed:  recordtest.P(recordpb.Grade_GRADE_LOW),
		}},
	}); err != nil {
		t.Fatal(err)
	}
	// The board must be otherwise CLEAN, or the open-gap arm answers first and the motion arm
	// — the one under test — is never reached.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: RoundIn(mustRun(t, runDir))("red-merge-r1")}, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("blue/report.md"),
		Prose:        proto.String("verified at the leaf"),
	}); err != nil {
		t.Fatal(err)
	}

	err := requirePassClosesAllGaps(mustRun(t, runDir))
	if err == nil {
		t.Fatal("PASS was allowed over an unruled motion")
	}
	if !strings.Contains(err.Error(), "show motions") {
		t.Errorf("the refusal does not name the read that unblocks it — handing a seat an id with "+
			"no way to look it up is what produced a ruling written blind: %v", err)
	}
}
