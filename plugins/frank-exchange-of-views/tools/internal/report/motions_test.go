package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// AN UNRULED MOTION IS SAID, not omitted. A filing with no answer means the sitting did not
// happen, and silence there reads identically to a run that never asked.
func TestAnUnruledMotionIsReported(t *testing.T) {
	b := famOf(t, []*record.Event{
		recordtest.Event(t, "blue-respond", &recordpb.Motion{
			MotionId: proto.String("M1"),
			Basis:    proto.String("the demand would bury a hazard"),
			Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
				Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
			}},
		}),
	})
	out := motions(b)
	if !strings.Contains(out, "NOT RULED") || !strings.Contains(out, "1 motion(s) received no ruling") {
		t.Errorf("an unanswered motion must be named, both on its row and in the tally:\n%s", out)
	}
}

// THE BENCH'S RULING ON THE CHAIR'S POSITION ON A WITHHELD PASS is one row with the position, and an
// upheld ruling's row names the work and the direction it owes.
func TestAPositionRulingRendersTheWorkItNames(t *testing.T) {
	p := record.Plan{PassPermitted: true, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{}, Exchanges: []record.ExchangeCount{}, Why: []string{}, StaleAreas: []record.StaleArea{}, Blockers: []record.PlanBlocker{}}
	for _, c := range []struct {
		name string
		h    *recordpb.PositionHearing
		want []string
	}{
		{"upheld on a lens", &recordpb.PositionHearing{Ruling: recordpb.PositionRuling_POSITION_RULING_UPHELD.Enum(), Lens: proto.String("red-lens-r1-logic"), Direction: proto.String("check the inference in §2")},
			[]string{"ruled upheld", "work named: lens red-lens-r1-logic — check the inference in §2"}},
		{"upheld on gaps", &recordpb.PositionHearing{Ruling: recordpb.PositionRuling_POSITION_RULING_UPHELD.Enum(), GapIds: []string{"G-00000001", "G-00000002"}, Direction: proto.String("source the figure")},
			[]string{"ruled upheld", "work named: gaps G-00000001, G-00000002, each sent back for one more exchange — source the figure"}},
		{"overruled", &recordpb.PositionHearing{Ruling: recordpb.PositionRuling_POSITION_RULING_OVERRULED.Enum()},
			[]string{"ruled overruled"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := famOf(t, []*record.Event{
				recordtest.Event(t, "red-chair", &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}),
				recordtest.Event(t, "red-chair", &recordpb.Position{PositionId: proto.String("A-0000000c"), Text: proto.String("the PASS is withheld")}),
				recordtest.Event(t, "judge", &recordpb.MotionRule{MotionId: proto.String("A-0000000c"), Opinion: proto.String("the bench's reasoning"),
					Ruling: &recordpb.MotionRule_Position{Position: c.h}}),
			})
			out := motions(b)
			for _, w := range append(c.want, "A-0000000c** · the chair's position on a PASS it withheld", "asked: the PASS is withheld", "the bench's reasoning") {
				if !strings.Contains(out, w) {
					t.Errorf("the motions section lacks %q:\n%s", w, out)
				}
			}
			if c.h.GetRuling() == recordpb.PositionRuling_POSITION_RULING_OVERRULED && strings.Contains(out, "work named") {
				t.Errorf("an overruled position names no work:\n%s", out)
			}
		})
	}
}

// A POSITION'S CLASS IS ON ITS ROW, with the material default that binds the hearing: the bench
// reads there that a position over narration of the run is upheld, never overruled (gblock,
// 2026-11-05).
func TestAPositionsClassRendersOnItsRow(t *testing.T) {
	p := record.Plan{PassPermitted: true, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{}, Exchanges: []record.ExchangeCount{}, Why: []string{}, StaleAreas: []record.StaleArea{}, Blockers: []record.PlanBlocker{}}
	b := famOf(t, []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}),
		recordtest.Event(t, "red-chair", &recordpb.Position{PositionId: proto.String("A-0000000c"), Text: proto.String("two sentences narrate the run"),
			Class: recordpb.GapClass_GAP_CLASS_REPORT_VOICE.Enum()}),
	})
	if out := motions(b); !strings.Contains(out, "A-0000000c** · the chair's position on a PASS it withheld, over a defect of class report_voice (material: always)") {
		t.Errorf("the position's row does not name its class and materiality:\n%s", out)
	}
}

// A WRITER WITH NO SITTING IS NAMED WITHOUT ONE. The harness files the docket motion migrate
// translates from a bench opinion and sits in no sitting, so its row names it bare; a seat's row
// keeps its sitting.
func TestAFilerWithNoSittingIsNamedWithoutANumber(t *testing.T) {
	var unruled int
	row := motionRow(&record.Motion{ID: "M1", Subject: "docket", GapID: "G1", Filer: record.HarnessSeat,
		Ruling: "repaired", RulingBy: "judge", RulingSitting: 2}, &unruled)
	if !strings.Contains(row, "filed by "+record.HarnessSeat+"\n") || strings.Contains(row, "#0") {
		t.Errorf("the harness's filing is named with a sitting number:\n%s", row)
	}
	if !strings.Contains(row, "by judge #2") {
		t.Errorf("the bench's ruling lost its sitting:\n%s", row)
	}
}
