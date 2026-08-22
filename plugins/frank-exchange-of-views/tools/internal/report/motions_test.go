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
	b := &record.Board{Events: []*record.Event{
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.Motion{
			MotionId: proto.String("M1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
			Basis:    proto.String("the demand would bury a hazard"),
			Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
				Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
			}},
		}),
	}}
	out := motions(b)
	if !strings.Contains(out, "NOT RULED") || !strings.Contains(out, "1 motion(s) received no ruling") {
		t.Errorf("an unanswered motion must be named, both on its row and in the tally:\n%s", out)
	}
}
