package scorecard

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func logEntry(t *testing.T, key string) *record.Event {
	t.Helper()
	return recordtest.At(t, "blue-respond", key, &recordpb.Log{
		Type:   recordpb.LogType_LOG_TYPE_FRICTION.Enum(),
		Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum(),
		Text:   proto.String("the changes projection paged awkwardly for this gap"),
	})
}

// THE LOG DUTY IS DISCHARGED BY THE REPAIR THAT FILES IT (#1002, gblock 2026-09-18). This metric
// ATTRIBUTES: every bucket is a sitting acts are filed into, and a sitting-record repair's acts are
// the repaired sitting's. So the repair opens no bucket of its own and the entry it files closes the
// sitting that owed it. Counted as a turn instead, one repair produces two false findings — the
// repaired sitting scored as having closed nothing, and the repair charged for a duty no act of its
// own can discharge — on the detector built to catch a real one.
func TestARepairsLogEntryClosesTheSittingItCompletes(t *testing.T) {
	repaired := []*record.Event{
		blueDispatch(t, "G1"),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, record.HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String("blue-respond:register:#1")}),
		logEntry(t, "blue-respond:log:l1"),
	}
	rows := blueRows(record.Run{}, nil, nil, famOfEventsT(repaired), record.WhileRunning)
	if r := rowByMetric(rows, "channel_closure"); r == nil || r.Value != 1.0 {
		t.Errorf("one sitting, its log filed by the repair that exists to file it = %+v, want 1", r)
	}
	if r := rowByMetric(rows, "sittings_never_closed"); r == nil || r.Value != 0 {
		t.Errorf("no sitting went unlogged = %+v, want 0", r)
	}

	// The control: the same seat sitting TWICE for real, logging once. Two buckets, one closed —
	// the reading this metric exists for, and the one a repair must not be confused with.
	twice := []*record.Event{
		blueDispatch(t, "G1"),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		logEntry(t, "blue-respond:log:l1"),
		blueDispatch(t, "G2"),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b")}),
	}
	rows = blueRows(record.Run{}, nil, nil, famOfEventsT(twice), record.WhileRunning)
	if r := rowByMetric(rows, "channel_closure"); r == nil || r.Value != 0.5 {
		t.Errorf("two sittings, one logged = %+v, want 0.5", r)
	}
	if r := rowByMetric(rows, "sittings_never_closed"); r == nil || r.Value != 1 {
		t.Errorf("one sitting closed no entry = %+v, want 1", r)
	}
}

// THE NUMERATOR IS AN INTERSECTION, SO THE METRIC CANNOT REPORT THE IMPOSSIBLE (#1026). Two set
// sizes divided is a ratio only while every closed bucket is also a sat bucket, and nothing made
// that true: a log entry filed by a seat with no register of its own buckets under a key `sat`
// never holds. Measured on the old division: channel_closure 2 and sittings_never_closed -1 — a
// closure rate above 1, and a detector whose sign nobody can act on.
//
// It also makes KEY IDENTITY load-bearing rather than incidental, which is what the fixture above
// could not see with one element in each set: bucket `closed` by turns while `sat` buckets by acts
// and a repair's entry joins nothing, which reads as 0 instead of silently reading as 1.
func TestAnEntryOutsideEverySittingCannotPushTheRatioAboveOne(t *testing.T) {
	evs := []*record.Event{
		blueDispatch(t, "G1"),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		logEntry(t, "blue-respond:log:l1"),
		// A seat that never registered on this stream, closing the channel anyway.
		recordtest.At(t, "red-lens-logic", "red-lens-logic:log:l1", &recordpb.Log{
			Type:   recordpb.LogType_LOG_TYPE_REQUEST.Enum(),
			Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum(),
			Text:   proto.String("no verb renders a gap's lineage in one read"),
		}),
	}
	rows := blueRows(record.Run{}, nil, nil, famOfEventsT(evs), record.WhileRunning)
	if r := rowByMetric(rows, "channel_closure"); r == nil || r.Value != 1.0 {
		t.Errorf("one sitting, its log filed, plus an entry belonging to no sitting = %+v, want 1", r)
	}
	if r := rowByMetric(rows, "sittings_never_closed"); r == nil || r.Value != 0 {
		t.Errorf("no sitting went unlogged = %+v, want 0 — a detector must never go negative", r)
	}
}

// A REGISTER OPENS THE BUCKET AND A LOG ENTRY CLOSES IT — NOTHING ELSE DOES EITHER (#1042). The two
// fixtures above reason about which events make a bucket but hold only the keys, so the arms
// themselves were free: adding `case EVENT_TYPE_POSITION: sat[key] = true` beside the register arm
// passed the package, and the metric then reads a denominator inflated by every act a seat files —
// half the channels scored as never closed, on a dashboard nobody re-derives. The same hole is open
// on the other arm: any act type added beside LOG discharges a duty no operator entry paid.
//
// The fixture holds both arms at once. The chair's position and the harness's sitting-close are
// acts by seats with no register on this stream, so opening a bucket on them ADDS one (0.5 becomes
// a third). Blue's first sitting files a position and an edit and NO entry, so closing a bucket on
// either act discharges it (0.5 becomes 1). An act by a registered seat inside its own sitting
// carries the key its register already opened, so it cannot inflate the denominator — the acts that
// can are exactly the ones here.
func TestOnlyARegisterOpensASittingBucketAndOnlyAnEntryCloses(t *testing.T) {
	evs := []*record.Event{
		blueDispatch(t, "G1"),
		recordtest.At(t, "red-chair", "red-chair:position:p1", &recordpb.Position{Text: proto.String("G1 stands")}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:position:p1", &recordpb.Position{Text: proto.String("G1 is repaired")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, record.HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		blueDispatch(t, "G2"),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b")}),
		logEntry(t, "blue-respond:log:l1"),
	}
	rows := blueRows(record.Run{}, nil, nil, famOfEventsT(evs), record.WhileRunning)
	if r := rowByMetric(rows, "channel_closure"); r == nil || r.Value != 0.5 {
		t.Errorf("two registers open two buckets and one entry closes one of them = %+v, want 0.5 — an act that is not a register opens no bucket, and an act that is not a log entry closes none", r)
	}
	if r := rowByMetric(rows, "sittings_never_closed"); r == nil || r.Value != 1 {
		t.Errorf("one of the two sittings filed no entry = %+v, want 1 — the position and the edit it did file discharge nothing", r)
	}
}
