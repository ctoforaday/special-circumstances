package scorecard

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func blueDispatch(t *testing.T, gaps ...string) *record.Event {
	return recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: gaps})
}

func lensCloses(t *testing.T, gap string) *record.Event {
	return recordtest.Event(t, "red-lens-logic", &recordpb.Close{GapId: proto.String(gap)})
}

func blueRegisters(t *testing.T) *record.Event {
	return recordtest.Event(t, "blue-respond", &recordpb.Register{AgentId: proto.String("blue-agent")})
}

// blueReturns is blue's agent's stop, which closes the sitting its register opened.
func blueReturns(t *testing.T) *record.Event {
	return recordtest.Event(t, record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("blue-agent"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")})
}

func blueRepairs(t *testing.T, gap string) *record.Event {
	return recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String(gap)})
}

func manifestRow(t *testing.T, gap string) *record.Event {
	return recordtest.Event(t, "blue-respond", &recordpb.ManifestRow{GapId: proto.String(gap), Row: proto.String("recomputed")})
}

// "SCORED AT CAPTURE" WAS SCORED BY NOTHING. The engine logs a gap with no row and defers it here;
// the denominator here was `repaired_gaps` summed from envelopes that never carried it, so every
// run scored a bare count and an owed-but-missing row was invisible. The denominator is now the
// record's owed set: repaired by blue's edit in a sitting that found the gap open (#868).
func TestManifestCoverageDividesByTheGapsBlueOwedOnTheRecord(t *testing.T) {
	t.Run("a repair with no row is scored missing, and named", func(t *testing.T) {
		fam := famOfEventsT([]*record.Event{
			blueDispatch(t, "G1", "G2"), blueRegisters(t), blueRepairs(t, "G1"), blueRepairs(t, "G2"), manifestRow(t, "G1"), blueReturns(t),
		})
		r := rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.WhileRunning), "manifest_coverage")
		if r == nil || r.Value != 0.5 {
			t.Fatalf("one of two owed gaps manifested must score 0.5, got %+v", r)
		}
		if !strings.Contains(r.Note, "G2") {
			t.Errorf("the owed-but-missing gap must be named: %q", r.Note)
		}
	})
	t.Run("the B4 ordering: gaps closed before blue sat are not owed, so no ratio charges blue", func(t *testing.T) {
		fam := famOfEventsT([]*record.Event{
			blueDispatch(t, "G1", "G2"), lensCloses(t, "G1"), lensCloses(t, "G2"), blueRegisters(t),
			// Edits naming gaps already closed when blue sat repair nothing that was open.
			blueRepairs(t, "G1"), blueRepairs(t, "G2"), blueReturns(t),
		})
		r := rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.WhileRunning), "manifest_coverage")
		if r == nil {
			t.Fatal("manifest_coverage row missing")
		}
		if v, isRatio := r.Value.(float64); isRatio {
			t.Errorf("nothing was owed, yet the row scored a ratio of %v", v)
		}
		if !strings.Contains(r.Note, "no gap was owed") {
			t.Errorf("the count fallback must say why it is not a ratio: %q", r.Note)
		}
	})
	t.Run("a closed-first gap and a rebutted gap both leave the denominator", func(t *testing.T) {
		fam := famOfEventsT([]*record.Event{
			blueDispatch(t, "G1", "G2", "G3"), lensCloses(t, "G1"), blueRegisters(t), blueRepairs(t, "G2"), manifestRow(t, "G2"), blueReturns(t),
		})
		r := rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.WhileRunning), "manifest_coverage")
		if r == nil || r.Value != 1.0 || r.Note != "" {
			t.Errorf("G2 was the only gap repaired while open and it carries a row — full coverage; G3 was rebutted and owes none; got %+v", r)
		}
	})
	t.Run("a sitting the record cannot close is not measured, never 'no gap was owed'", func(t *testing.T) {
		fam := famOfEventsT([]*record.Event{blueDispatch(t, "G1"), blueRegisters(t)})
		r := rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.WhileRunning), "manifest_coverage")
		if r == nil || strings.Contains(r.Note, "no gap was owed") || !strings.Contains(r.Note, "1 blue sitting(s) NOT MEASURED") {
			t.Errorf("blue has neither registered again nor returned, so what it owed is not known: %+v", r)
		}
		fam = famOfEventsT([]*record.Event{blueDispatch(t, "G1"), blueRegisters(t), blueRepairs(t, "G1")})
		r = rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.WhileRunning), "manifest_coverage")
		if r == nil || r.Value != 0.0 || !strings.Contains(r.Note, "1 of 1 owed gap(s) carry no row: G1") || !strings.Contains(r.Note, "NOT MEASURED") {
			t.Errorf("the owed gap is scored and the unclosed sitting is named beside it: %+v", r)
		}
		// AFTER THE RUN THE END OF THE RECORD CLOSES IT: capture scores the same record with nothing
		// left in flight, so the count is the answer and nothing is not measured.
		r = rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.AfterTheRun), "manifest_coverage")
		if r == nil || r.Value != 0.0 || r.Note != "1 of 1 owed gap(s) carry no row: G1" {
			t.Errorf("after the run the sitting is closed and only the missing row is noted: %+v", r)
		}
		fam = famOfEventsT([]*record.Event{blueDispatch(t, "G1"), blueRegisters(t)})
		r = rowByMetric(blueRows(record.Run{}, nil, nil, fam, record.AfterTheRun), "manifest_coverage")
		if r == nil || strings.Contains(r.Note, "NOT MEASURED") || !strings.Contains(r.Note, "no gap was owed") {
			t.Errorf("after the run a sitting that repaired nothing owed nothing: %+v", r)
		}
	})
	t.Run("an unreadable record is not measured, never a zero", func(t *testing.T) {
		r := rowByMetric(blueRows(record.Run{}, nil, nil, nil, record.WhileRunning), "manifest_coverage")
		if r == nil || r.Value != nil || !strings.Contains(r.Note, "not measured") {
			t.Errorf("a nil record must say not measured, got %+v", r)
		}
	})
}
