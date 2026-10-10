package view

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE DEBATE OVER A DIRECTION LIVES IN THE DIRECTIONS' OWN DOCUMENT. Red's ruling is an argument
// beside the line, and what blue did about it is the line's own path: a reader following one
// avenue meets the ruling, its opinion, and the move blue made with that ruling in front of it.
func TestAnAvenueRulingRendersBesideTheMoveMadeAgainstIt(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Line: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "red-chair", &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE),
			Opinion:  proto.String("a real question, not this one's"),
			Ruling:   &recordpb.MotionRule_Avenue{Avenue: recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE},
		}),
		recordtest.Event(t, "blue-r1", &recordpb.Avenue{AvenueId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Reason: proto.String("the adjacent literature is what the question turns on")}),
	}
	got := string(avenueMD(inputT(t, evs...)))
	for _, want := range []string{
		"RED RULED **out_of_scope**",
		"a real question, not this one's",
		"the adjacent literature is what the question turns on",
		"proposed -> e0 pursued",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the avenue's document does not carry %q:\n%s", want, got)
		}
	}
}
