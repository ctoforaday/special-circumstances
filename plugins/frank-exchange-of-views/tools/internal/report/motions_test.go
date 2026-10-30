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
