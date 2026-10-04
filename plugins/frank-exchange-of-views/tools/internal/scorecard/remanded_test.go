package scorecard

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// Blue's position prose is written once and kept, so a run recorded while the deferring
// disposition was spelled "carried" still says so. Direction uptake reads that prose: a pattern
// that knew only the current word would score such a run's uptake as zero, silently.
func TestDirectionUptakeReadsBothSpellingsOfTheDeferral(t *testing.T) {
	for _, prose := range []string{"the gap was remanded; acting on the stated research", "the gap was carried; acting on the stated research"} {
		dj := record.DebateJSON{Epochs: []record.DebateEpochJSON{
			{Lead: []record.DebateOpinionJSON{{GapID: "G1"}}, Blue: []string{prose}},
		}}
		if lead, blue := ComputeDirectionUptake(dj); lead != 1 || blue != 1 {
			t.Errorf("%q: uptake %d/%d, want 1/1", prose, blue, lead)
		}
	}
}

// benchSatOn seeds one docket motion per disposition, each ruled by the bench with a principle
// and an opinion — the shape the write requires of every docket ruling.
func benchSatOn(t *testing.T, opinions []string, ds ...recordpb.Disposition) *record.Family {
	t.Helper()
	var evs []*record.Event
	for i, d := range ds {
		id := fmt.Sprintf("M%d", i+1)
		evs = append(evs, docketFiling(t, id, fmt.Sprintf("G%d", i+1)))
		opinion := "heard"
		if i < len(opinions) {
			opinion = opinions[i]
		}
		evs = append(evs, docketRuling(t, id, d, opinion))
	}
	return famSeededT(t, evs)
}

func docketFiling(t *testing.T, motion, gap string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Motion{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}},
	})
}

func docketRuling(t *testing.T, motion string, d recordpb.Disposition, opinion string) *record.Event {
	t.Helper()
	r := &recordpb.DocketRuling{
		Disposition: recordtest.P(d), Principle: proto.String("a ruling states its principle"),
		Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"),
	}
	// A ruling states what reopens the gap or that it is final, never both.
	if d == recordpb.Disposition_DISPOSITION_REMANDED {
		r.ReopensOn = proto.String("the stated direction reporting back")
	} else {
		r.Final = proto.Bool(true)
	}
	return recordtest.Event(t, "judge", &recordpb.MotionRule{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Opinion:  proto.String(opinion),
		Ruling:   &recordpb.MotionRule_Docket{Docket: r},
	})
}

// THE BENCH CARD READS THE RECORD. The journal's judge envelopes restate the dispositions and carry
// no opinion, so the card counted what an envelope happened to restate (4 of 5 rulings on m11, a
// remanded_share of 0.5 against the record's 0.6) and scored every ruling opinionless. The journal
// passed here is empty: every figure must come off the record.
func TestTheBenchCardReadsTheRecord(t *testing.T) {
	rem, rep := recordpb.Disposition_DISPOSITION_REMANDED, recordpb.Disposition_DISPOSITION_REPAIRED
	fam := benchSatOn(t, []string{"read off the trajectory of the lens's tool calls"}, rem, rem, rem, rep, rep)
	rows := benchRows(fam)

	t.Run("remanded_share", func(t *testing.T) {
		r := rowByMetric(rows, "remanded_share")
		if r == nil {
			t.Fatal("no remanded_share row")
		}
		if v, ok := r.Value.(float64); !ok || v != 0.6 {
			t.Errorf("remanded_share = %v (%s), want 0.6 — 3 of the record's 5 docket rulings", r.Value, r.Note)
		}
		if !strings.HasPrefix(r.Note, "3/5") {
			t.Errorf("remanded_share note = %q, want it to state 3/5", r.Note)
		}
	})
	// A DETECTOR THE WRITE MAKES IMPOSSIBLE MEASURES NOTHING. Every docket ruling carries a
	// principle — the write refuses one without — so the row could only ever print 0 off the
	// record, and off the journal it printed every ruling.
	t.Run("rulings_without_opinion is gone", func(t *testing.T) {
		if r := rowByMetric(rows, "rulings_without_opinion"); r != nil {
			t.Errorf("rulings_without_opinion is on the bench card: %+v", *r)
		}
	})
	t.Run("undeclared_inspection_risk reads the recorded opinions", func(t *testing.T) {
		r := rowByMetric(rows, "undeclared_inspection_risk")
		if r == nil {
			t.Fatal("no undeclared_inspection_risk row")
		}
		if !strings.HasPrefix(r.Note, "1 opinion(s) reference trajectory evidence") {
			t.Errorf("undeclared_inspection_risk note = %q, want the one opinion that names the trajectory", r.Note)
		}
	})
}

// An unread record is not a bench that never sat: the rows say they were not measured.
func TestTheBenchCardOnAnUnreadRecordIsNotMeasured(t *testing.T) {
	rows := benchRows(nil)
	for _, metric := range []string{"remanded_share", "undeclared_inspection_risk"} {
		r := rowByMetric(rows, metric)
		if r == nil {
			t.Fatalf("no %s row", metric)
		}
		if r.Value != nil || !strings.Contains(r.Note, "not measured") {
			t.Errorf("%s on an unread record = %v (%q), want no value and a not-measured note", metric, r.Value, r.Note)
		}
	}
}

func TestABenchThatDidNotSitSaysSo(t *testing.T) {
	r := rowByMetric(benchRows(famOfEventsT(nil)), "remanded_share")
	if r == nil || r.Value != nil || r.Note != "the bench did not sit this run" {
		t.Errorf("remanded_share with no docket rulings = %+v, want the did-not-sit note", r)
	}
}

// ONE RULING PER MOTION, ON BOTH ROWS. A motion the bench ruled in one sitting and ruled again in
// a later one is one ruling — the one that stands. remanded_share and undeclared_inspection_risk
// count over the same rulings, so the opinion share and the remand share have one denominator.
func TestAMotionRuledInTwoSittingsIsOneRuling(t *testing.T) {
	rem, rep := recordpb.Disposition_DISPOSITION_REMANDED, recordpb.Disposition_DISPOSITION_REPAIRED
	const traj = "read off the trajectory of the lens's tool calls"
	fam := famSeededT(t, []*record.Event{
		recordtest.Event(t, "judge", &recordpb.Register{AgentId: proto.String("J1"), Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}),
		docketFiling(t, "M1", "G1"),
		docketRuling(t, "M1", rem, traj),
		docketFiling(t, "M2", "G2"),
		docketRuling(t, "M2", rem, "heard"),
		recordtest.Event(t, "judge", &recordpb.Register{AgentId: proto.String("J2"), Occasion: recordpb.Occasion_OCCASION_TERMINAL.Enum()}),
		docketRuling(t, "M1", rep, traj),
	})
	rows := benchRows(fam)
	if r := rowByMetric(rows, "remanded_share"); r == nil || r.Value != 0.5 || !strings.HasPrefix(r.Note, "1/2") {
		t.Errorf("remanded_share = %+v, want 0.5 at 1/2: M1 stands repaired, M2 remanded", r)
	}
	if r := rowByMetric(rows, "undeclared_inspection_risk"); r == nil || !strings.HasPrefix(r.Note, "1 opinion(s)") {
		t.Errorf("undeclared_inspection_risk = %+v, want 1 opinion: M1 is one ruling however often it was ruled", r)
	}
}

// The principle is half the opinion: a ruling whose principle declares the inspection is counted
// though its argument does not.
func TestAPrincipleThatDeclaresTheInspectionCounts(t *testing.T) {
	fam := famSeededT(t, []*record.Event{
		docketFiling(t, "M1", "G1"),
		recordtest.Event(t, "judge", &recordpb.MotionRule{
			MotionId: proto.String("M1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion:  proto.String("heard"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
				Principle:   proto.String("a claim about a tool call is checked against the tool call"),
				Tension:     proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"), Final: proto.Bool(true),
			}},
		}),
	})
	if r := rowByMetric(benchRows(fam), "undeclared_inspection_risk"); r == nil || !strings.HasPrefix(r.Note, "1 opinion(s)") {
		t.Errorf("undeclared_inspection_risk = %+v, want the one ruling whose principle names the tool call", r)
	}
}
