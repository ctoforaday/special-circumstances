package view

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE DEBATE OVER A DIRECTION LIVES IN THE DIRECTIONS' OWN DOCUMENT. The report printed red's
// ruling and blue's appeal under the line until they were ruled debate rather than subject. The
// ruling already rendered here; the appeal rendered in judgments.md (with its reason, among every
// motion) and now also here, beside the ruling it answers, where a reader following one line meets it.
func TestAnAppealRendersBesideItsRuling(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Line: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "red-chair", &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION),
			Opinion:  proto.String("a real question, not this one's"),
			Ruling:   &recordpb.MotionRule_Direction{Direction: recordpb.DirectionRuling_DIRECTION_RULING_OUT_OF_SCOPE},
		}),
		recordtest.Event(t, "blue-r1", &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "blue-r1", &recordpb.MotionAppeal{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION),
			Reason:   proto.String("the adjacent literature is what the question turns on"),
		}),
	}
	got := string(inquiryMD(Input{Events: evs}))
	ruled := strings.Index(got, "RED RULED **out_of_scope**")
	appealed := strings.Index(got, "BLUE APPEALED the `out_of_scope` ruling")
	if ruled < 0 || appealed < 0 || appealed < ruled {
		t.Errorf("the ruling and the appeal against it must both render, appeal after ruling:\n%s", got)
	}
	if !strings.Contains(got, "a real question, not this one's") {
		t.Errorf("the ruling's opinion must render with it:\n%s", got)
	}
}
