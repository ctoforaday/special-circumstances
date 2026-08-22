package graph

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A hole is a genuine defect: an unanswered dispute, or a torn closure (closed with no reason).
// A merge close that carries its closure_class is NOT a hole even with no opinion — the false
// positive this test pins down (R1-7 in the 2026-07-22 run flagged amber until this was fixed).
func TestGapHoleHeuristic(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"MERGE_CLOSED", "TORN", "UNANSWERED", "OPEN"},
		Gaps: map[string]*record.Gap{
			// A closure that STATES its class is not a hole. The fixture said
			// `evidence-rebutted`, which is not a ClosureClass and never was — untyped it read as
			// a closure with a class, and typed it cannot be written at all.
			"MERGE_CLOSED": {ID: "MERGE_CLOSED", Open: false, Mint: mint(), Closure: &recordpb.Close{
				ClosureClass: recordtest.P(recordpb.ClosureClass_CLOSURE_CLASS_REBUTTAL_SUSTAINED),
			}},
			// TORN is a closure carrying NO class — which is the whole condition, and which the
			// schema expresses by absence rather than by an unrecognised word.
			"TORN":       {ID: "TORN", Open: false, Mint: mint(), Closure: &recordpb.Close{}},
			"UNANSWERED": {ID: "UNANSWERED", Open: true, Mint: mint()},
			"OPEN":       {ID: "OPEN", Open: true, Mint: mint()},
		},
		Events: []*record.Event{
			// A GRADE MOTION FILED AND NEVER RULED. This fixture used a `dispute` event until the
			// motion collapse retired the type; the counters then read zero for every run and the
			// hole detector could not fire at all, while this test went on passing against a
			// vocabulary nothing wrote.
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Motion{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("UNANSWERED")}},
			}),
		},
	}
	m := gapFlowMermaid(b)

	// A merge close with a closure_class is closed, NOT a hole.
	if lineClass(m, "g_MERGE_CLOSED") != "closed" {
		t.Errorf("a merge-close with closure_class must be 'closed', not a hole:\n%s", m)
	}
	// A closed gap with no reason is a torn closure -> hole.
	if lineClass(m, "g_TORN") != "hole" {
		t.Errorf("a torn closure (no reason) must be a hole:\n%s", m)
	}
	// An open gap with a filed-but-unruled grade motion is a hole.
	if lineClass(m, "g_UNANSWERED") != "hole" {
		t.Errorf("a grade motion with no ruling must be a hole:\n%s", m)
	}
	// A plain open gap is open, not a hole.
	if lineClass(m, "g_OPEN") != "open" {
		t.Errorf("a plain open gap must be 'open':\n%s", m)
	}
}

// lineClass finds the `:::class` suffix on the node's line.
func lineClass(mermaid, node string) string {
	for _, ln := range strings.Split(mermaid, "\n") {
		if strings.Contains(ln, node+"[") {
			if i := strings.LastIndex(ln, ":::"); i >= 0 {
				return strings.TrimSpace(ln[i+3:])
			}
		}
	}
	return ""
}

func TestSeatFlowTalliesEvents(t *testing.T) {
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.Mint{}),
		},
	}
	m := seatFlowMermaid(b)
	if !strings.Contains(m, "mint×2") || !strings.Contains(m, "red-merge-r1") {
		t.Errorf("seat flow should tally events per seat:\n%s", m)
	}
}

// A RULED grade motion is not a hole — the counter must distinguish filed-and-answered from
// filed-and-ignored, which is the entire reason it counts two numbers rather than one.
func TestRuledGradeMotionIsNotAHole(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"ANSWERED"},
		Gaps:     map[string]*record.Gap{"ANSWERED": {ID: "ANSWERED", Open: true, Mint: mint()}},
		Events: []*record.Event{
			recordtest.Event(t, "blue-respond-r1", 1, &recordpb.Motion{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("ANSWERED")}},
			}),
			// THE RULING IS THE ONEOF ARM. `accepted` can only reach the grade arm, so a fixture
			// cannot pair a grade motion with a petition's verdict the way two loose strings could.
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.MotionRule{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_ACCEPTED},
			}),
		},
	}
	if got := lineClass(gapFlowMermaid(b), "g_ANSWERED"); got != "open" {
		t.Errorf("a ruled grade motion must not be a hole, got %q:\n%s", got, gapFlowMermaid(b))
	}
}

// mint is the minimum a gap needs to exist on the board for these tests: the class, and nothing
// else. Every case here is about CLOSURE state, so the mint carries no grades — an absent grade is
// a real state and inventing one would make each fixture say more than the test means.
func mint() *recordpb.Mint {
	return &recordpb.Mint{Class: proto.String("c")}
}
